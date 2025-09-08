package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/streed/ml-notes/internal/autotag"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/search"
	"github.com/streed/ml-notes/internal/summarize"
)

// AssetProvider interface for accessing web assets
type AssetProvider interface {
	GetTemplates() (*template.Template, error)
	GetStaticHandler() http.Handler
	HasEmbeddedAssets() bool
}

type APIServer struct {
	cfg          *config.Config
	db           *sql.DB
	repo         *models.NoteRepository
	vectorSearch search.SearchProvider
	autoTagger   *autotag.AutoTagger
	server       *http.Server
	templates    *template.Template
	assets       AssetProvider
	useEmbedded  bool
	webDir       string // fallback for development
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type SearchRequest struct {
	Query     string `json:"query"`
	Tags      string `json:"tags"`
	Limit     int    `json:"limit"`
	UseVector bool   `json:"use_vector"`
}

type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags"`
	AutoTag bool   `json:"auto_tag"`
}

type UpdateNoteRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Tags    string `json:"tags,omitempty"`
}

type AutoTagRequest struct {
	NoteIDs   []int `json:"note_ids,omitempty"`
	All       bool  `json:"all"`
	Recent    int   `json:"recent"`
	Apply     bool  `json:"apply"`
	Overwrite bool  `json:"overwrite"`
}

type UpdateSettingsRequest struct {
	OllamaEndpoint      string `json:"ollama_endpoint,omitempty"`
	Debug               *bool  `json:"debug,omitempty"`
	SummarizationModel  string `json:"summarization_model,omitempty"`
	EnableSummarization *bool  `json:"enable_summarization,omitempty"`
	Editor              string `json:"editor,omitempty"`
	EnableAutoTagging   *bool  `json:"enable_auto_tagging,omitempty"`
	MaxAutoTags         *int   `json:"max_auto_tags,omitempty"`
	GitHubOwner         string `json:"github_owner,omitempty"`
	GitHubRepo          string `json:"github_repo,omitempty"`
}

func NewAPIServer(cfg *config.Config, db *sql.DB, repo *models.NoteRepository, vectorSearch search.SearchProvider, assetProvider AssetProvider) *APIServer {
	var templates *template.Template
	useEmbedded := false
	webDir := ""

	// Try to load assets from provider first
	if assetProvider != nil && assetProvider.HasEmbeddedAssets() {
		var err error
		templates, err = assetProvider.GetTemplates()
		if err != nil {
			logger.Debug("Failed to load embedded templates: %v, falling back to filesystem", err)
		} else {
			logger.Debug("Loaded embedded templates successfully")
			useEmbedded = true
		}
	}

	// Fallback to filesystem if embedded assets failed
	if !useEmbedded {
		webDir = findWebAssetsDir()
		if webDir != "" {
			templatePath := filepath.Join(webDir, "templates", "*.html")
			var err error
			templates, err = template.ParseGlob(templatePath)
			if err != nil {
				logger.Debug("Failed to load templates from %s: %v (web UI will be disabled)", templatePath, err)
			} else {
				logger.Debug("Loaded templates from %s", templatePath)
			}
		} else {
			logger.Debug("Web assets directory not found (web UI will be disabled)")
		}
	}

	return &APIServer{
		cfg:          cfg,
		db:           db,
		repo:         repo,
		vectorSearch: vectorSearch,
		autoTagger:   autotag.NewAutoTagger(cfg),
		templates:    templates,
		assets:       assetProvider,
		useEmbedded:  useEmbedded,
		webDir:       webDir,
	}
}

// findWebAssetsDir looks for the web directory in several locations
func findWebAssetsDir() string {
	// Try current directory first
	if dirExists("web") {
		return "web"
	}

	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		webPath := filepath.Join(execDir, "web")
		if dirExists(webPath) {
			return webPath
		}

		// Try one level up from executable (common during development)
		webPath = filepath.Join(execDir, "..", "web")
		if dirExists(webPath) {
			return webPath
		}
	}

	// Try some common paths
	commonPaths := []string{
		"/usr/share/ml-notes/web",
		"/opt/ml-notes/web",
		"./web",
	}

	for _, path := range commonPaths {
		if dirExists(path) {
			return path
		}
	}

	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (s *APIServer) Start(host string, port int) error {
	router := mux.NewRouter()

	// Web UI routes (if templates are available)
	if s.templates != nil {
		router.HandleFunc("/", s.handleWebUI).Methods("GET")
		router.HandleFunc("/new", s.handleNewNote).Methods("GET")
		router.HandleFunc("/note/{id:[0-9]+}", s.handleWebNote).Methods("GET")
		router.HandleFunc("/graph", s.handleGraphUI).Methods("GET")
		router.HandleFunc("/settings", s.handleSettingsUI).Methods("GET")

		// Serve static files - embedded or filesystem
		if s.useEmbedded && s.assets != nil {
			// Serve embedded static assets
			router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", s.assets.GetStaticHandler()))
			logger.Debug("Web UI enabled, serving embedded static assets")
		} else if s.webDir != "" {
			// Serve static files from filesystem
			staticDir := filepath.Join(s.webDir, "static")
			router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
			logger.Debug("Web UI enabled, serving static files from %s", staticDir)
		}

		// Custom CSS route (only for filesystem mode)
		if !s.useEmbedded && s.cfg.WebUICustomCSS != "" {
			router.HandleFunc("/static/css/custom.css", s.handleCustomCSS).Methods("GET")
		}
	}

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Notes endpoints
	api.HandleFunc("/notes", s.handleListNotes).Methods("GET")
	api.HandleFunc("/notes", s.handleCreateNote).Methods("POST")
	api.HandleFunc("/notes/search", s.handleSearchNotes).Methods("POST")
	api.HandleFunc("/notes/{id:[0-9]+}", s.handleGetNote).Methods("GET")
	api.HandleFunc("/notes/{id:[0-9]+}", s.handleUpdateNote).Methods("PUT")
	api.HandleFunc("/notes/{id:[0-9]+}", s.handleDeleteNote).Methods("DELETE")

	// Tags endpoints
	api.HandleFunc("/tags", s.handleListTags).Methods("GET")
	api.HandleFunc("/notes/{id:[0-9]+}/tags", s.handleUpdateNoteTags).Methods("PUT")

	// Auto-tagging endpoints
	api.HandleFunc("/auto-tag/suggest/{id:[0-9]+}", s.handleSuggestTags).Methods("POST")
	api.HandleFunc("/auto-tag/apply", s.handleAutoTag).Methods("POST")

	// Analysis endpoints
	api.HandleFunc("/analyze/{id:[0-9]+}", s.handleAnalyzeNote).Methods("POST")

	// Graph visualization endpoint
	api.HandleFunc("/graph", s.handleGraphData).Methods("GET")

	// Statistics and info endpoints
	api.HandleFunc("/stats", s.handleStats).Methods("GET")
	api.HandleFunc("/config", s.handleConfig).Methods("GET")
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Settings endpoints
	api.HandleFunc("/settings", s.handleGetSettings).Methods("GET")
	api.HandleFunc("/settings", s.handleUpdateSettings).Methods("POST")
	api.HandleFunc("/settings/test-ollama", s.handleTestOllama).Methods("POST")

	// Serve OpenAPI documentation
	api.HandleFunc("/docs", s.handleDocs).Methods("GET")

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure this more restrictively in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	})

	handler := c.Handler(router)

	addr := fmt.Sprintf("%s:%d", host, port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting HTTP API server on %s", addr)
	return s.server.ListenAndServe()
}

func (s *APIServer) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *APIServer) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: statusCode < 400,
		Data:    data,
	}

	if err, ok := data.(error); ok {
		response.Success = false
		response.Error = err.Error()
		response.Data = nil
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode JSON response: %v", err)
	}
}

func (s *APIServer) writeError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: false,
		Error:   err.Error(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode JSON response: %v", err)
	}
}

func (s *APIServer) parseIntParam(r *http.Request, param string) (int, error) {
	vars := mux.Vars(r)
	str, exists := vars[param]
	if !exists {
		return 0, fmt.Errorf("missing parameter: %s", param)
	}
	return strconv.Atoi(str)
}

func (s *APIServer) parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}

	var tags []string
	for _, tag := range strings.Split(tagsStr, ",") {
		cleanTag := strings.TrimSpace(tag)
		if cleanTag != "" {
			tags = append(tags, cleanTag)
		}
	}
	return tags
}

// Handlers

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":       "ok",
		"timestamp":    time.Now().Format(time.RFC3339),
		"version":      "1.0.0",
		"auto_tagging": s.cfg.EnableAutoTagging,
	}

	// Check database connection
	if err := s.db.Ping(); err != nil {
		health["status"] = "unhealthy"
		health["database_error"] = err.Error()
		s.writeJSON(w, http.StatusServiceUnavailable, health)
		return
	}

	// Check auto-tagging availability
	if s.cfg.EnableAutoTagging {
		health["auto_tagging_available"] = s.autoTagger.IsAvailable()
	}

	s.writeJSON(w, http.StatusOK, health)
}

func (s *APIServer) handleListNotes(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	notes, err := s.repo.ListWithLimit(limit, offset)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, notes)
}

func (s *APIServer) handleGetNote(w http.ResponseWriter, r *http.Request) {
	id, err := s.parseIntParam(r, "id")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	note, err := s.repo.GetByID(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	s.writeJSON(w, http.StatusOK, note)
}

func (s *APIServer) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if req.Title == "" || req.Content == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("title and content are required"))
		return
	}

	// Parse initial tags
	tags := s.parseTags(req.Tags)

	// Auto-tag if requested
	if req.AutoTag && s.cfg.EnableAutoTagging && s.autoTagger.IsAvailable() {
		tempNote := &models.Note{
			Title:   req.Title,
			Content: req.Content,
		}

		if suggestedTags, err := s.autoTagger.SuggestTags(tempNote); err == nil {
			// Merge with existing tags
			tagSet := make(map[string]bool)
			for _, tag := range tags {
				tagSet[tag] = true
			}
			for _, tag := range suggestedTags {
				if !tagSet[tag] {
					tags = append(tags, tag)
				}
			}
		}
	}

	var note *models.Note
	var err error

	if len(tags) > 0 {
		note, err = s.repo.CreateWithTags(req.Title, req.Content, tags)
	} else {
		note, err = s.repo.Create(req.Title, req.Content)
	}

	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	// Index for vector search
	if s.vectorSearch != nil {
		fullText := note.Title + " " + note.Content
		// Use namespace for indexing if it's LilRagSearch
		if lilragSearch, ok := s.vectorSearch.(*search.LilRagSearch); ok {
			namespace := s.getCurrentProjectNamespace()
			if err := lilragSearch.IndexNoteWithNamespace(note.ID, fullText, namespace); err != nil {
				logger.Error("Failed to index note %d: %v", note.ID, err)
			}
		} else {
			if err := s.vectorSearch.IndexNote(note.ID, fullText); err != nil {
				logger.Error("Failed to index note %d: %v", note.ID, err)
			}
		}
	}

	s.writeJSON(w, http.StatusCreated, note)
}

func (s *APIServer) handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	id, err := s.parseIntParam(r, "id")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	// Get existing note
	note, err := s.repo.GetByID(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	// Update fields if provided
	if req.Title != "" {
		note.Title = req.Title
	}
	if req.Content != "" {
		note.Content = req.Content
	}

	// Update tags if provided
	if req.Tags != "" {
		tags := s.parseTags(req.Tags)
		if err := s.repo.UpdateTags(note.ID, tags); err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	// Update in database
	if err := s.repo.Update(note); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	// Re-index for vector search
	if s.vectorSearch != nil {
		fullText := note.Title + " " + note.Content
		// Use namespace for indexing if it's LilRagSearch
		if lilragSearch, ok := s.vectorSearch.(*search.LilRagSearch); ok {
			namespace := s.getCurrentProjectNamespace()
			if err := lilragSearch.IndexNoteWithNamespace(note.ID, fullText, namespace); err != nil {
				logger.Error("Failed to re-index note %d: %v", note.ID, err)
			}
		} else {
			if err := s.vectorSearch.IndexNote(note.ID, fullText); err != nil {
				logger.Error("Failed to re-index note %d: %v", note.ID, err)
			}
		}
	}

	// Get updated note
	updatedNote, err := s.repo.GetByID(id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, updatedNote)
}

func (s *APIServer) handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := s.parseIntParam(r, "id")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.repo.Delete(id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Note deleted successfully"})
}

func (s *APIServer) handleSearchNotes(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if req.Query == "" && req.Tags == "" {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("query or tags must be provided"))
		return
	}

	// Set default limit
	if req.Limit == 0 {
		if req.UseVector && s.vectorSearch != nil {
			req.Limit = 1
		} else {
			req.Limit = 10
		}
	}

	var notes []*models.Note
	var err error

	// Handle tag search
	if req.Tags != "" {
		tags := s.parseTags(req.Tags)
		notes, err = s.repo.SearchByTags(tags)
	} else if req.UseVector && s.vectorSearch != nil {
		// Use namespace for vector search if it's LilRagSearch
		if lilragSearch, ok := s.vectorSearch.(*search.LilRagSearch); ok {
			namespace := s.getCurrentProjectNamespace()
			notes, err = lilragSearch.SearchSimilarWithNamespace(req.Query, req.Limit, namespace)
		} else {
			notes, err = s.vectorSearch.SearchSimilar(req.Query, req.Limit)
		}
	} else {
		notes, err = s.repo.Search(req.Query)
	}

	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	// Limit results if text search returned too many
	if len(notes) > req.Limit {
		notes = notes[:req.Limit]
	}

	s.writeJSON(w, http.StatusOK, notes)
}

func (s *APIServer) handleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.repo.GetAllTags()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, tags)
}

func (s *APIServer) handleUpdateNoteTags(w http.ResponseWriter, r *http.Request) {
	id, err := s.parseIntParam(r, "id")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		Tags string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	// Verify note exists
	_, err = s.repo.GetByID(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	tags := s.parseTags(req.Tags)
	if err := s.repo.UpdateTags(id, tags); err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Tags updated successfully",
		"tags":    tags,
	})
}

func (s *APIServer) handleSuggestTags(w http.ResponseWriter, r *http.Request) {
	id, err := s.parseIntParam(r, "id")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	// Check if auto-tagging is available
	if !s.autoTagger.IsAvailable() {
		s.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("auto-tagging is not available"))
		return
	}

	// Get the note
	note, err := s.repo.GetByID(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	// Get tag suggestions
	suggestedTags, err := s.autoTagger.SuggestTags(note)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"note_id":        id,
		"note_title":     note.Title,
		"suggested_tags": suggestedTags,
		"existing_tags":  note.Tags,
	})
}

func (s *APIServer) handleAutoTag(w http.ResponseWriter, r *http.Request) {
	var req AutoTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	// Check if auto-tagging is available
	if !s.autoTagger.IsAvailable() {
		s.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("auto-tagging is not available"))
		return
	}

	// Determine which notes to process
	var notes []*models.Note
	var err error

	if req.All {
		notes, err = s.repo.List(0, 0) // Get all notes
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}
	} else if req.Recent > 0 {
		notes, err = s.repo.List(req.Recent, 0)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}
	} else if len(req.NoteIDs) > 0 {
		for _, id := range req.NoteIDs {
			note, noteErr := s.repo.GetByID(id)
			if noteErr != nil {
				s.writeError(w, http.StatusNotFound, fmt.Errorf("note %d not found: %w", id, noteErr))
				return
			}
			notes = append(notes, note)
		}
	} else {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("must specify note_ids, all=true, or recent > 0"))
		return
	}

	if len(notes) == 0 {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "No notes found to process",
			"results": []interface{}{},
		})
		return
	}

	// Process notes for auto-tagging
	results := []map[string]interface{}{}
	successCount := 0
	errorCount := 0

	for _, note := range notes {
		result := map[string]interface{}{
			"note_id":    note.ID,
			"note_title": note.Title,
			"success":    false,
		}

		// Get suggested tags
		suggestedTags, err := s.autoTagger.SuggestTags(note)
		if err != nil {
			result["error"] = err.Error()
			errorCount++
			results = append(results, result)
			continue
		}

		if len(suggestedTags) == 0 {
			result["message"] = "No tags suggested"
			results = append(results, result)
			continue
		}

		result["suggested_tags"] = suggestedTags
		result["existing_tags"] = note.Tags

		// Determine final tags
		var finalTags []string
		if req.Overwrite || len(note.Tags) == 0 {
			finalTags = suggestedTags
		} else {
			// Merge with existing tags
			tagSet := make(map[string]bool)
			for _, tag := range note.Tags {
				tagSet[tag] = true
				finalTags = append(finalTags, tag)
			}
			for _, tag := range suggestedTags {
				if !tagSet[tag] {
					finalTags = append(finalTags, tag)
				}
			}
		}

		result["final_tags"] = finalTags

		// Apply tags if requested
		if req.Apply {
			if err := s.repo.UpdateTags(note.ID, finalTags); err != nil {
				result["error"] = fmt.Sprintf("Failed to apply tags: %v", err)
				errorCount++
			} else {
				result["success"] = true
				result["applied"] = true
				successCount++
			}
		} else {
			result["success"] = true
			result["applied"] = false
			successCount++
		}

		results = append(results, result)
	}

	response := map[string]interface{}{
		"processed_count": len(notes),
		"success_count":   successCount,
		"error_count":     errorCount,
		"applied":         req.Apply,
		"results":         results,
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleStats(w http.ResponseWriter, r *http.Request) {
	// Get total count
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	// Get tag count
	var tagCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM tags").Scan(&tagCount)
	if err != nil {
		tagCount = 0 // Fallback if tags table doesn't exist
	}

	stats := map[string]interface{}{
		"total_notes":   count,
		"total_tags":    tagCount,
		"auto_tagging":  s.cfg.EnableAutoTagging,
		"database_path": s.cfg.GetDatabasePath(),
	}

	s.writeJSON(w, http.StatusOK, stats)
}

func (s *APIServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"debug_mode":           s.cfg.Debug,
		"auto_tagging_enabled": s.cfg.EnableAutoTagging,
		"max_auto_tags":        s.cfg.MaxAutoTags,
		"data_directory":       s.cfg.DataDirectory,
	}

	s.writeJSON(w, http.StatusOK, config)
}

func (s *APIServer) handleDocs(w http.ResponseWriter, r *http.Request) {
	docs := `# ML Notes API Documentation

## Base URL
http://localhost:8080/api/v1

## Endpoints

### Notes
- GET /notes - List notes (query params: limit, offset)
- GET /notes/{id} - Get specific note
- POST /notes - Create new note
- PUT /notes/{id} - Update note
- DELETE /notes/{id} - Delete note
- POST /notes/search - Search notes

### Tags
- GET /tags - List all tags
- PUT /notes/{id}/tags - Update note tags

### Auto-tagging
- POST /auto-tag/suggest/{id} - Suggest tags for note
- POST /auto-tag/apply - Apply auto-tagging to notes

### System
- GET /health - Health check
- GET /stats - Database statistics
- GET /config - Configuration info
- GET /docs - This documentation

## Example Usage

### Create a note with auto-tagging:
POST /notes
{
  "title": "My Note",
  "content": "Note content here",
  "auto_tag": true
}

### Search notes:
POST /notes/search
{
  "query": "search term",
  "limit": 10,
  "use_vector": true
}

### Auto-tag multiple notes:
POST /auto-tag/apply
{
  "note_ids": [1, 2, 3],
  "apply": true,
  "overwrite": false
}
`

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(docs)); err != nil {
		logger.Error("Failed to write response: %v", err)
	}
}

// Web UI handlers

func (s *APIServer) handleWebUI(w http.ResponseWriter, r *http.Request) {
	// Get recent notes for the sidebar
	notes, err := s.repo.List(50, 0)
	if err != nil {
		logger.Error("Failed to load notes for web UI: %v", err)
		http.Error(w, "Failed to load notes", http.StatusInternalServerError)
		return
	}

	// Get tags for filtering
	tags, err := s.repo.GetAllTags()
	if err != nil {
		logger.Error("Failed to load tags for web UI: %v", err)
		tags = []models.Tag{} // Fallback to empty
	}

	data := map[string]interface{}{
		"Config": s.cfg,
		"Notes":  notes,
		"Tags":   tags,
		"Stats": map[string]interface{}{
			"TotalNotes":  len(notes),
			"AutoTagging": s.cfg.EnableAutoTagging,
		},
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		logger.Error("Failed to render template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func (s *APIServer) handleWebNote(w http.ResponseWriter, r *http.Request) {
	id, err := s.parseIntParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	note, err := s.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	// Get recent notes for the sidebar
	notes, err := s.repo.List(50, 0)
	if err != nil {
		logger.Error("Failed to load notes for web UI: %v", err)
		notes = []*models.Note{} // Fallback to empty
	}

	// Get tags for filtering
	tags, err := s.repo.GetAllTags()
	if err != nil {
		logger.Error("Failed to load tags for web UI: %v", err)
		tags = []models.Tag{} // Fallback to empty
	}

	data := map[string]interface{}{
		"Config":      s.cfg,
		"Notes":       notes,
		"Tags":        tags,
		"CurrentNote": note,
		"Stats": map[string]interface{}{
			"TotalNotes":  len(notes),
			"AutoTagging": s.cfg.EnableAutoTagging,
		},
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		logger.Error("Failed to render template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
func (s *APIServer) handleAnalyzeNote(w http.ResponseWriter, r *http.Request) {
	id, err := s.parseIntParam(r, "id")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	// Check if summarization is enabled
	if !s.cfg.EnableSummarization {
		s.writeError(w, http.StatusServiceUnavailable, fmt.Errorf("analysis is disabled in configuration"))
		return
	}

	// Get the note
	note, err := s.repo.GetByID(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	// Parse query parameters
	writeBack := r.URL.Query().Get("write-back") == "true"
	writeNew := r.URL.Query().Get("write-new") == "true"
	writeTitle := r.URL.Query().Get("write-title")
	prompt := r.URL.Query().Get("prompt")

	// Validate parameters
	if writeBack && writeNew {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("cannot use both write-back and write-new options"))
		return
	}

	// Create analyzer
	analyzer := summarize.NewSummarizer(s.cfg)
	if s.cfg.SummarizationModel != "" {
		analyzer.SetModel(s.cfg.SummarizationModel)
	}

	// Generate analysis
	result, err := analyzer.SummarizeNoteWithPrompt(note, prompt)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate analysis: %w", err))
		return
	}

	response := map[string]interface{}{
		"analysis":        result.Summary,
		"model":           result.Model,
		"original_length": result.OriginalLength,
		"summary_length":  result.SummaryLength,
		"compression":     100.0 * (1.0 - float64(result.SummaryLength)/float64(result.OriginalLength)),
	}

	// Handle write-back to current note
	if writeBack {
		analysisSection := fmt.Sprintf("\n\n---\n## Analysis\n\n%s\n\n*Analysis generated on %s using %s*",
			result.Summary,
			time.Now().Format("2006-01-02 15:04:05"),
			result.Model)

		newContent := note.Content + analysisSection
		_, err := s.repo.UpdateByID(note.ID, note.Title, newContent)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to update note with analysis: %w", err))
			return
		}
		response["written_back"] = true
	}

	// Handle write to new note
	if writeNew {
		newTitle := writeTitle
		if newTitle == "" {
			newTitle = fmt.Sprintf("Analysis of %s", note.Title)
		}

		sourceInfo := fmt.Sprintf("**Source:** Note %d - %s", note.ID, note.Title)
		newContent := fmt.Sprintf("%s\n\n---\n\n%s\n\n---\n\n*Analysis generated on %s using %s*",
			sourceInfo,
			result.Summary,
			time.Now().Format("2006-01-02 15:04:05"),
			result.Model)

		newNote, err := s.repo.Create(newTitle, newContent)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to create new note with analysis: %w", err))
			return
		}

		// Index the note for vector search
		if s.vectorSearch != nil {
			fullText := newTitle + " " + newContent
			// Use namespace for indexing if it's LilRagSearch
			if lilragSearch, ok := s.vectorSearch.(*search.LilRagSearch); ok {
				namespace := s.getCurrentProjectNamespace()
				if err := lilragSearch.IndexNoteWithNamespace(newNote.ID, fullText, namespace); err != nil {
					logger.Error("Failed to index analysis note %d: %v", newNote.ID, err)
				}
			} else {
				if err := s.vectorSearch.IndexNote(newNote.ID, fullText); err != nil {
					logger.Error("Failed to index analysis note %d: %v", newNote.ID, err)
				}
			}
		}

		response["new_note_id"] = newNote.ID
		response["new_note_title"] = newTitle
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleCustomCSS(w http.ResponseWriter, r *http.Request) {
	if s.cfg.WebUICustomCSS == "" {
		http.NotFound(w, r)
		return
	}

	// Serve the custom CSS file
	http.ServeFile(w, r, s.cfg.WebUICustomCSS)
}

func (s *APIServer) handleNewNote(w http.ResponseWriter, r *http.Request) {
	// Get recent notes for the sidebar
	notes, err := s.repo.List(50, 0)
	if err != nil {
		logger.Error("Failed to load notes for web UI: %v", err)
		notes = []*models.Note{} // Fallback to empty
	}

	// Get tags for filtering
	tags, err := s.repo.GetAllTags()
	if err != nil {
		logger.Error("Failed to load tags for web UI: %v", err)
		tags = []models.Tag{} // Fallback to empty
	}

	// Create a new empty note object for the editor
	newNote := &models.Note{
		ID:      0, // 0 indicates new note
		Title:   "",
		Content: "",
		Tags:    []string{},
	}

	data := map[string]interface{}{
		"Config":      s.cfg,
		"Notes":       notes,
		"Tags":        tags,
		"CurrentNote": newNote,
		"IsNewNote":   true,
		"Stats": map[string]interface{}{
			"TotalNotes":  len(notes),
			"AutoTagging": s.cfg.EnableAutoTagging,
		},
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		logger.Error("Failed to render template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// Graph data structures
type GraphNode struct {
	ID    int      `json:"id"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
	Size  int      `json:"size"`  // Based on content length or connections
	Group int      `json:"group"` // For coloring based on primary tag
}

type GraphEdge struct {
	Source     int      `json:"source"`
	Target     int      `json:"target"`
	Weight     float64  `json:"weight"`      // Strength of connection (0-1)
	SharedTags []string `json:"shared_tags"` // The actual shared tags
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphEdge `json:"links"`
}

func (s *APIServer) handleGraphData(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	minConnections := 1
	if minStr := r.URL.Query().Get("min_connections"); minStr != "" {
		if m, err := strconv.Atoi(minStr); err == nil && m > 0 {
			minConnections = m
		}
	}

	maxNodes := 100
	if maxStr := r.URL.Query().Get("max_nodes"); maxStr != "" {
		if m, err := strconv.Atoi(maxStr); err == nil && m > 0 {
			maxNodes = m
		}
	}

	tagFilter := r.URL.Query().Get("tag_filter")

	// Get all notes with tags
	notes, err := s.repo.List(0, 0) // Get all notes
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	// Filter notes that have tags and optionally by tag filter
	var filteredNotes []*models.Note
	for _, note := range notes {
		if len(note.Tags) == 0 {
			continue // Skip notes without tags
		}

		// Apply tag filter if specified
		if tagFilter != "" {
			hasTag := false
			for _, tag := range note.Tags {
				if strings.Contains(strings.ToLower(tag), strings.ToLower(tagFilter)) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		filteredNotes = append(filteredNotes, note)
	}

	// Limit nodes if too many
	if len(filteredNotes) > maxNodes {
		filteredNotes = filteredNotes[:maxNodes]
	}

	// Build graph data
	graphData := s.buildGraphData(filteredNotes, minConnections)

	s.writeJSON(w, http.StatusOK, graphData)
}

func (s *APIServer) buildGraphData(notes []*models.Note, minConnections int) *GraphData {
	// Create nodes
	nodes := make([]GraphNode, 0, len(notes))
	noteMap := make(map[int]*models.Note)

	// Get tag frequency for grouping
	tagCount := make(map[string]int)
	for _, note := range notes {
		noteMap[note.ID] = note
		for _, tag := range note.Tags {
			tagCount[tag]++
		}
	}

	// Find most common tags for grouping
	tagToGroup := make(map[string]int)
	groupCounter := 0
	for tag, count := range tagCount {
		if count >= 2 { // Only group tags that appear in multiple notes
			tagToGroup[tag] = groupCounter
			groupCounter++
		}
	}

	for _, note := range notes {
		// Determine group based on most frequent tag
		group := 0
		for _, tag := range note.Tags {
			if g, exists := tagToGroup[tag]; exists {
				group = g
				break
			}
		}

		node := GraphNode{
			ID:    note.ID,
			Title: note.Title,
			Tags:  note.Tags,
			Size:  len(note.Content)/10 + 10, // Base size + content length factor
			Group: group,
		}
		nodes = append(nodes, node)
	}

	// Create edges based on shared tags
	edges := make([]GraphEdge, 0)
	connectionCount := make(map[int]int) // Track connections per node

	for i, note1 := range notes {
		for j, note2 := range notes {
			if i >= j { // Avoid duplicate edges and self-loops
				continue
			}

			// Find shared tags
			sharedTags := s.findSharedTags(note1.Tags, note2.Tags)
			if len(sharedTags) == 0 {
				continue
			}

			// Calculate connection strength (Jaccard similarity)
			weight := s.calculateTagSimilarity(note1.Tags, note2.Tags)

			edge := GraphEdge{
				Source:     note1.ID,
				Target:     note2.ID,
				Weight:     weight,
				SharedTags: sharedTags,
			}
			edges = append(edges, edge)

			connectionCount[note1.ID]++
			connectionCount[note2.ID]++
		}
	}

	// Filter out nodes with too few connections if requested
	if minConnections > 1 {
		var filteredNodes []GraphNode
		var filteredEdges []GraphEdge
		validNodes := make(map[int]bool)

		// Identify nodes with enough connections
		for _, node := range nodes {
			if connectionCount[node.ID] >= minConnections {
				validNodes[node.ID] = true
				filteredNodes = append(filteredNodes, node)
			}
		}

		// Keep only edges between valid nodes
		for _, edge := range edges {
			if validNodes[edge.Source] && validNodes[edge.Target] {
				filteredEdges = append(filteredEdges, edge)
			}
		}

		nodes = filteredNodes
		edges = filteredEdges
	}

	return &GraphData{
		Nodes: nodes,
		Links: edges,
	}
}

func (s *APIServer) findSharedTags(tags1, tags2 []string) []string {
	tagSet := make(map[string]bool)
	for _, tag := range tags1 {
		tagSet[tag] = true
	}

	var shared []string
	for _, tag := range tags2 {
		if tagSet[tag] {
			shared = append(shared, tag)
		}
	}

	return shared
}

// getCurrentProjectNamespace gets the current working directory name as project namespace
func (s *APIServer) getCurrentProjectNamespace() string {
	wd, err := os.Getwd()
	if err != nil {
		logger.Debug("Failed to get working directory for namespace: %v", err)
		return ""
	}
	return filepath.Base(wd)
}

func (s *APIServer) calculateTagSimilarity(tags1, tags2 []string) float64 {
	if len(tags1) == 0 && len(tags2) == 0 {
		return 0.0
	}

	// Create sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, tag := range tags1 {
		set1[tag] = true
	}
	for _, tag := range tags2 {
		set2[tag] = true
	}

	// Calculate intersection and union
	intersection := 0
	union := make(map[string]bool)

	for tag := range set1 {
		union[tag] = true
		if set2[tag] {
			intersection++
		}
	}
	for tag := range set2 {
		union[tag] = true
	}

	// Jaccard similarity: |intersection| / |union|
	if len(union) == 0 {
		return 0.0
	}

	return float64(intersection) / float64(len(union))
}

func (s *APIServer) handleGraphUI(w http.ResponseWriter, r *http.Request) {
	// Get tags for filtering
	tags, err := s.repo.GetAllTags()
	if err != nil {
		logger.Error("Failed to load tags for graph UI: %v", err)
		tags = []models.Tag{} // Fallback to empty
	}

	data := map[string]interface{}{
		"Config": s.cfg,
		"Tags":   tags,
		"Stats": map[string]interface{}{
			"AutoTagging": s.cfg.EnableAutoTagging,
		},
	}

	if err := s.templates.ExecuteTemplate(w, "graph.html", data); err != nil {
		logger.Error("Failed to render graph template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// Settings handlers

func (s *APIServer) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings := map[string]interface{}{
		"data_directory":       s.cfg.DataDirectory,
		"database_path":        s.cfg.GetDatabasePath(),
		"ollama_endpoint":      s.cfg.OllamaEndpoint,
		"debug":                s.cfg.Debug,
		"summarization_model":  s.cfg.SummarizationModel,
		"enable_summarization": s.cfg.EnableSummarization,
		"editor":               s.cfg.Editor,
		"enable_auto_tagging":  s.cfg.EnableAutoTagging,
		"max_auto_tags":        s.cfg.MaxAutoTags,
		"github_owner":         s.cfg.GitHubOwner,
		"github_repo":          s.cfg.GitHubRepo,
		"webui_theme":          s.cfg.WebUITheme,
	}

	s.writeJSON(w, http.StatusOK, settings)
}

func (s *APIServer) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	// Create a new config copy to update
	newCfg := *s.cfg
	// Update fields if provided
	if req.OllamaEndpoint != "" {
		newCfg.OllamaEndpoint = req.OllamaEndpoint
	}
	if req.Debug != nil {
		newCfg.Debug = *req.Debug
	}
	if req.SummarizationModel != "" {
		newCfg.SummarizationModel = req.SummarizationModel
	}
	if req.EnableSummarization != nil {
		newCfg.EnableSummarization = *req.EnableSummarization
	}
	if req.Editor != "" {
		newCfg.Editor = req.Editor
	}
	if req.EnableAutoTagging != nil {
		newCfg.EnableAutoTagging = *req.EnableAutoTagging
	}
	if req.MaxAutoTags != nil {
		if *req.MaxAutoTags < 1 || *req.MaxAutoTags > 20 {
			s.writeError(w, http.StatusBadRequest, fmt.Errorf("max auto tags must be between 1 and 20"))
			return
		}
		newCfg.MaxAutoTags = *req.MaxAutoTags
	}
	if req.GitHubOwner != "" {
		newCfg.GitHubOwner = req.GitHubOwner
	}
	if req.GitHubRepo != "" {
		newCfg.GitHubRepo = req.GitHubRepo
	}

	// Save the updated configuration
	if err := config.Save(&newCfg); err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to save configuration: %w", err))
		return
	}

	// Update the server's config
	s.cfg = &newCfg

	// Check if vector configuration changed
	response := map[string]interface{}{
		"message": "Settings updated successfully",
		"settings": map[string]interface{}{
			"data_directory":       s.cfg.DataDirectory,
			"database_path":        s.cfg.GetDatabasePath(),
			"ollama_endpoint":      s.cfg.OllamaEndpoint,
			"debug":                s.cfg.Debug,
			"summarization_model":  s.cfg.SummarizationModel,
			"enable_summarization": s.cfg.EnableSummarization,
			"editor":               s.cfg.Editor,
			"enable_auto_tagging":  s.cfg.EnableAutoTagging,
			"max_auto_tags":        s.cfg.MaxAutoTags,
			"github_owner":         s.cfg.GitHubOwner,
			"github_repo":          s.cfg.GitHubRepo,
			"webui_theme":          s.cfg.WebUITheme,
		},
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleTestOllama(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	// Use provided endpoint or current config
	endpoint := req.Endpoint
	if endpoint == "" {
		endpoint = s.cfg.OllamaEndpoint
	}
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Test Ollama connection by fetching models
	testURL := endpoint + "/api/tags"
	httpReq, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to create request: %w", err))
		return
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		// Determine specific error type
		var errorMsg string
		if strings.Contains(err.Error(), "timeout") {
			errorMsg = "Connection timeout (10s)"
		} else if strings.Contains(err.Error(), "refused") {
			errorMsg = "Connection refused - is Ollama running?"
		} else if strings.Contains(err.Error(), "no such host") {
			errorMsg = "Host not found - check the URL"
		} else {
			errorMsg = err.Error()
		}

		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":    false,
			"error":      errorMsg,
			"endpoint":   endpoint,
			"tested_url": testURL,
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":    false,
			"error":      fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status),
			"endpoint":   endpoint,
			"tested_url": testURL,
		})
		return
	}

	// Try to parse the response to count models
	var ollamaResp struct {
		Models []map[string]interface{} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		// Connection worked but couldn't parse response
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success":    true,
			"message":    "Connection successful (could not parse model list)",
			"endpoint":   endpoint,
			"tested_url": testURL,
		})
		return
	}

	// Success with model count
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("Connection successful! Found %d models", len(ollamaResp.Models)),
		"model_count": len(ollamaResp.Models),
		"endpoint":    endpoint,
		"tested_url":  testURL,
	})
}

func (s *APIServer) handleSettingsUI(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Config": s.cfg,
		"Stats": map[string]interface{}{
			"AutoTagging": s.cfg.EnableAutoTagging,
		},
	}

	if err := s.templates.ExecuteTemplate(w, "settings.html", data); err != nil {
		logger.Error("Failed to render settings template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
