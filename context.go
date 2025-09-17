package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/models"
)

// Notes API methods for Wails frontend

// GetNote retrieves a note by ID
func (a *App) GetNote(id int) (*models.Note, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}
	return a.services.Notes.GetByID(id)
}

// ListNotes retrieves notes with pagination
func (a *App) ListNotes(limit, offset int) ([]*models.Note, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}
	if limit <= 0 {
		limit = 50
	}
	return a.services.Notes.List(limit, offset)
}

// CreateNote creates a new note
func (a *App) CreateNote(title, content string, tags []string) (*models.Note, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	return a.services.Notes.Create(title, content, tags)
}

// UpdateNote updates an existing note
func (a *App) UpdateNote(id int, title, content string, tags []string) (*models.Note, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	// Get existing note
	note, err := a.services.Notes.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	// Update fields
	note.Title = title
	note.Content = content

	// Update tags if provided
	if tags != nil {
		if err := a.services.Tags.UpdateNoteTags(id, tags); err != nil {
			return nil, fmt.Errorf("failed to update tags: %w", err)
		}
	}

	// Update note
	if err := a.services.Notes.Update(note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// Return updated note
	return a.services.Notes.GetByID(id)
}

// DeleteNote deletes a note by ID
func (a *App) DeleteNote(id int) error {
	if a.services == nil {
		return fmt.Errorf("services not initialized")
	}
	return a.services.Notes.Delete(id)
}

// SearchNotes searches for notes
func (a *App) SearchNotes(query string, useVector bool, limit int) ([]*models.Note, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}
	if limit <= 0 {
		limit = 10
	}
	return a.services.Search.SearchNotes(query, useVector, limit)
}

// Tags API methods

// GetAllTags retrieves all tags
func (a *App) GetAllTags() ([]models.Tag, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}
	return a.services.Tags.GetAll()
}

// UpdateNoteTags updates tags for a specific note
func (a *App) UpdateNoteTags(noteID int, tags []string) error {
	if a.services == nil {
		return fmt.Errorf("services not initialized")
	}
	return a.services.Tags.UpdateNoteTags(noteID, tags)
}

// SearchByTags searches notes by tags
func (a *App) SearchByTags(tags []string) ([]*models.Note, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}
	return a.services.Search.SearchByTags(tags)
}

// Auto-tagging API methods

// IsAutoTagAvailable checks if auto-tagging is available
func (a *App) IsAutoTagAvailable() bool {
	if a.services == nil {
		return false
	}
	return a.services.AutoTag.IsAvailable()
}

// SuggestTags suggests tags for a note
func (a *App) SuggestTags(noteID int) ([]string, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	note, err := a.services.Notes.GetByID(noteID)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	return a.services.AutoTag.SuggestTags(note)
}

// Analysis API methods

// AnalyzeNote analyzes a note with optional custom prompt
func (a *App) AnalyzeNote(noteID int, prompt string) (map[string]interface{}, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	note, err := a.services.Notes.GetByID(noteID)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	result, err := a.services.Analyze.AnalyzeNote(note, prompt)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	return map[string]interface{}{
		"analysis":        result.Summary,
		"model":           result.Model,
		"original_length": result.OriginalLength,
		"summary_length":  result.SummaryLength,
		"compression":     100.0 * (1.0 - float64(result.SummaryLength)/float64(result.OriginalLength)),
	}, nil
}

// Preferences API methods

// GetPreference gets a string preference
func (a *App) GetPreference(key, defaultValue string) string {
	if a.services == nil {
		return defaultValue
	}
	return a.services.Preferences.GetString(key, defaultValue)
}

// SetPreference sets a string preference
func (a *App) SetPreference(key, value string) error {
	if a.services == nil {
		return fmt.Errorf("services not initialized")
	}
	return a.services.Preferences.SetString(key, value)
}

// GetBoolPreference gets a boolean preference
func (a *App) GetBoolPreference(key string, defaultValue bool) bool {
	if a.services == nil {
		return defaultValue
	}
	return a.services.Preferences.GetBool(key, defaultValue)
}

// SetBoolPreference sets a boolean preference
func (a *App) SetBoolPreference(key string, value bool) error {
	if a.services == nil {
		return fmt.Errorf("services not initialized")
	}
	return a.services.Preferences.SetBool(key, value)
}

// GetJSONPreference gets a JSON preference
func (a *App) GetJSONPreference(key string, target interface{}) error {
	if a.services == nil {
		return fmt.Errorf("services not initialized")
	}
	return a.services.Preferences.GetJSON(key, target)
}

// SetJSONPreference sets a JSON preference
func (a *App) SetJSONPreference(key string, value interface{}) error {
	if a.services == nil {
		return fmt.Errorf("services not initialized")
	}
	return a.services.Preferences.SetJSON(key, value)
}

// Utility methods

// GetStats returns basic application statistics
func (a *App) GetStats() (map[string]interface{}, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	notes, err := a.services.Notes.List(0, 0)
	if err != nil {
		return nil, err
	}

	tags, err := a.services.Tags.GetAll()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_notes":   len(notes),
		"total_tags":    len(tags),
		"auto_tagging":  a.services.AutoTag.IsAvailable(),
		"database_path": a.services.Config.GetDatabasePath(),
	}, nil
}

// ShowNotification shows a notification to the user (Wails runtime)
func (a *App) ShowNotification(title, message, notificationType string) {
	// This could use Wails runtime notification or custom modal
	// For now, we'll implement custom modals in the frontend
}

// ShowError shows an error dialog
func (a *App) ShowError(title, message string) {
	// This will be implemented with frontend modals
}

// ShowSuccess shows a success message
func (a *App) ShowSuccess(message string) {
	// This will be implemented with frontend modals
}

// Configuration API methods

// GetConfig returns the current configuration
func (a *App) GetConfig() (map[string]interface{}, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	cfg := a.services.Config
	return map[string]interface{}{
		"data_directory":       cfg.DataDirectory,
		"database_path":        cfg.GetDatabasePath(),
		"ollama_endpoint":      cfg.OllamaEndpoint,
		"debug":                cfg.Debug,
		"summarization_model":  cfg.SummarizationModel,
		"enable_summarization": cfg.EnableSummarization,
		"editor":               cfg.Editor,
		"enable_auto_tagging":  cfg.EnableAutoTagging,
		"auto_tag_model":       cfg.AutoTagModel,
		"max_auto_tags":        cfg.MaxAutoTags,
		"webui_theme":          cfg.WebUITheme,
		"webui_custom_css":     cfg.WebUICustomCSS,
		"github_owner":         cfg.GitHubOwner,
		"github_repo":          cfg.GitHubRepo,
		"lilrag_url":           cfg.LilRagURL,
		"current_project":      cfg.CurrentProject,
	}, nil
}

// UpdateConfig updates the configuration with provided values
func (a *App) UpdateConfig(updates map[string]interface{}) error {
	if a.services == nil {
		return fmt.Errorf("services not initialized")
	}

	cfg := a.services.Config

	// Update configuration fields
	for key, value := range updates {
		switch key {
		case "data_directory":
			if str, ok := value.(string); ok {
				cfg.DataDirectory = str
			}
		case "ollama_endpoint":
			if str, ok := value.(string); ok {
				cfg.OllamaEndpoint = str
			}
		case "debug":
			if b, ok := value.(bool); ok {
				cfg.Debug = b
			}
		case "summarization_model":
			if str, ok := value.(string); ok {
				cfg.SummarizationModel = str
			}
		case "enable_summarization":
			if b, ok := value.(bool); ok {
				cfg.EnableSummarization = b
			}
		case "editor":
			if str, ok := value.(string); ok {
				cfg.Editor = str
			}
		case "enable_auto_tagging":
			if b, ok := value.(bool); ok {
				cfg.EnableAutoTagging = b
			}
		case "auto_tag_model":
			if str, ok := value.(string); ok {
				cfg.AutoTagModel = str
			}
		case "max_auto_tags":
			if f, ok := value.(float64); ok {
				cfg.MaxAutoTags = int(f)
			}
		case "webui_theme":
			if str, ok := value.(string); ok {
				cfg.WebUITheme = str
			}
		case "webui_custom_css":
			if str, ok := value.(string); ok {
				cfg.WebUICustomCSS = str
			}
		case "lilrag_url":
			if str, ok := value.(string); ok {
				cfg.LilRagURL = str
			}
		}
	}

	// Save the updated configuration
	return config.Save(cfg)
}

// InitializeConfig initializes a new configuration
func (a *App) InitializeConfig(dataDir, ollamaEndpoint string) error {
	_, err := config.InitializeConfig(dataDir, ollamaEndpoint)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Restart services with new config (this would require app restart in practice)
	return nil
}

// IsConfigInitialized checks if the application has been properly configured
func (a *App) IsConfigInitialized() bool {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return !os.IsNotExist(err)
}

// TestOllamaConnection tests the connection to Ollama service
func (a *App) TestOllamaConnection() (map[string]interface{}, error) {
	if a.services == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	// Simple test by trying to connect to the Ollama endpoint
	cfg := a.services.Config
	endpoint := cfg.OllamaEndpoint + "/api/tags"

	resp, err := http.Get(endpoint)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log the error but don't fail the operation since we're testing connectivity
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode == 200 {
		return map[string]interface{}{
			"success": true,
			"message": "Successfully connected to Ollama",
		}, nil
	}

	return map[string]interface{}{
		"success": false,
		"error":   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
	}, nil
}
