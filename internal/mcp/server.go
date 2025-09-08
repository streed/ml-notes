package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/streed/ml-notes/internal/autotag"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/search"
)

type NotesServer struct {
	cfg          *config.Config
	db           *sql.DB
	repo         *models.NoteRepository
	vectorSearch search.SearchProvider
	mcpServer    *server.MCPServer
	autoTagger   *autotag.AutoTagger
}

func NewNotesServer(cfg *config.Config, db *sql.DB, repo *models.NoteRepository, vectorSearch search.SearchProvider) *NotesServer {
	ns := &NotesServer{
		cfg:          cfg,
		db:           db,
		repo:         repo,
		vectorSearch: vectorSearch,
		autoTagger:   autotag.NewAutoTagger(cfg),
	}

	// Create MCP server
	ns.mcpServer = server.NewMCPServer(
		"ml-notes",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	ns.registerTools()

	// Register resources
	ns.registerResources()

	// Register prompts
	ns.registerPrompts()

	return ns
}

func (s *NotesServer) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}

func (s *NotesServer) registerTools() {
	// Add note tool
	addNoteTool := mcp.NewTool("add_note",
		mcp.WithDescription("Add a new note to the database"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("The title of the note"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The content of the note"),
		),
		mcp.WithString("tags",
			mcp.Description("Comma-separated tags for the note (optional)"),
		),
	)
	s.mcpServer.AddTool(addNoteTool, s.handleAddNote)

	// Search notes tool
	searchTool := mcp.NewTool("search_notes",
		mcp.WithDescription("Search for notes using vector similarity, text search, or tag search. Vector search returns the single most similar note by default."),
		mcp.WithString("query",
			mcp.Description("Search query string (optional if tags are provided)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (default: 1 for vector, 10 for text/tags)"),
		),
		mcp.WithBoolean("use_vector",
			mcp.Description("Use vector search if available (default: true)"),
		),
		mcp.WithString("tags",
			mcp.Description("Comma-separated tags to search for (optional)"),
		),
	)
	s.mcpServer.AddTool(searchTool, s.handleSearchNotes)

	// Get note tool
	getNoteTool := mcp.NewTool("get_note",
		mcp.WithDescription("Get a specific note by ID"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("The ID of the note to retrieve"),
		),
	)
	s.mcpServer.AddTool(getNoteTool, s.handleGetNote)

	// List notes tool
	listNotesTool := mcp.NewTool("list_notes",
		mcp.WithDescription("List all notes with optional limit"),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of notes to return"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of notes to skip"),
		),
	)
	s.mcpServer.AddTool(listNotesTool, s.handleListNotes)

	// Update note tool
	updateNoteTool := mcp.NewTool("update_note",
		mcp.WithDescription("Update an existing note"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("The ID of the note to update"),
		),
		mcp.WithString("title",
			mcp.Description("New title for the note (optional)"),
		),
		mcp.WithString("content",
			mcp.Description("New content for the note (optional)"),
		),
		mcp.WithString("tags",
			mcp.Description("Comma-separated tags to set for the note (optional, replaces existing tags)"),
		),
	)
	s.mcpServer.AddTool(updateNoteTool, s.handleUpdateNote)

	// Delete note tool
	deleteNoteTool := mcp.NewTool("delete_note",
		mcp.WithDescription("Delete a note by ID"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("The ID of the note to delete"),
		),
	)
	s.mcpServer.AddTool(deleteNoteTool, s.handleDeleteNote)

	// List tags tool
	listTagsTool := mcp.NewTool("list_tags",
		mcp.WithDescription("List all tags in the system"),
	)
	s.mcpServer.AddTool(listTagsTool, s.handleListTags)

	// Update note tags tool
	updateTagsTool := mcp.NewTool("update_note_tags",
		mcp.WithDescription("Update the tags for a specific note"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("The ID of the note to update tags for"),
		),
		mcp.WithString("tags",
			mcp.Required(),
			mcp.Description("Comma-separated tags to set for the note (replaces existing tags)"),
		),
	)
	s.mcpServer.AddTool(updateTagsTool, s.handleUpdateNoteTags)

	// Auto-tag tools
	suggestTagsTool := mcp.NewTool("suggest_tags",
		mcp.WithDescription("Suggest tags for a note using AI analysis"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("The ID of the note to analyze for tag suggestions"),
		),
	)
	s.mcpServer.AddTool(suggestTagsTool, s.handleSuggestTags)

	autoTagTool := mcp.NewTool("auto_tag_note",
		mcp.WithDescription("Automatically generate and apply tags to a note using AI"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("The ID of the note to auto-tag"),
		),
		mcp.WithBoolean("overwrite",
			mcp.Description("Whether to overwrite existing tags (default: false, merges with existing)"),
		),
	)
	s.mcpServer.AddTool(autoTagTool, s.handleAutoTagNote)
}

func (s *NotesServer) registerResources() {
	// Recent notes resource
	recentResource := mcp.NewResource("notes://recent",
		"Recent Notes",
		mcp.WithResourceDescription("Get the most recently created notes"),
		mcp.WithMIMEType("application/json"),
	)
	s.mcpServer.AddResource(recentResource, s.handleRecentNotes)

	// Stats resource
	statsResource := mcp.NewResource("notes://stats",
		"Notes Statistics",
		mcp.WithResourceDescription("Get statistics about the notes database"),
		mcp.WithMIMEType("application/json"),
	)
	s.mcpServer.AddResource(statsResource, s.handleStats)

	// Config resource
	configResource := mcp.NewResource("notes://config",
		"Configuration",
		mcp.WithResourceDescription("Get current ml-notes configuration"),
		mcp.WithMIMEType("application/json"),
	)
	s.mcpServer.AddResource(configResource, s.handleConfig)
}

func (s *NotesServer) registerPrompts() {
	// Search notes prompt
	searchPrompt := mcp.NewPrompt("search_notes",
		mcp.WithPromptDescription("Search for notes using vector similarity or text search"),
		mcp.WithArgument("query",
			mcp.ArgumentDescription("Search query string"),
		),
		mcp.WithArgument("limit",
			mcp.ArgumentDescription("Maximum number of results (default: 10)"),
		),
	)
	s.mcpServer.AddPrompt(searchPrompt, s.handleSearchPrompt)

	// Summarize notes prompt
	summarizePrompt := mcp.NewPrompt("summarize_notes",
		mcp.WithPromptDescription("Generate a summary prompt for all notes"),
	)
	s.mcpServer.AddPrompt(summarizePrompt, s.handleSummarizePrompt)
}

// Tool handlers
func (s *NotesServer) handleAddNote(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: add_note")

	title, err := request.RequireString("title")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'title': %w", err)
	}

	content, err := request.RequireString("content")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'content': %w", err)
	}

	// Parse tags if provided
	tagsStr := request.GetString("tags", "")
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			cleanTag := strings.TrimSpace(tag)
			if cleanTag != "" {
				tags = append(tags, cleanTag)
			}
		}
	}

	var note *models.Note
	if len(tags) > 0 {
		note, err = s.repo.CreateWithTags(title, content, tags)
	} else {
		note, err = s.repo.Create(title, content)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Index for vector search
	if s.vectorSearch != nil {
		fullText := title + " " + content
		if err := s.vectorSearch.IndexNote(note.ID, fullText); err != nil {
			logger.Error("Failed to index note %d: %v", note.ID, err)
		}
	}

	result := fmt.Sprintf("Note created successfully with ID: %d\nTitle: %s", note.ID, note.Title)
	if len(note.Tags) > 0 {
		result += fmt.Sprintf("\nTags: %s", strings.Join(note.Tags, ", "))
	}
	return mcp.NewToolResultText(result), nil
}

func (s *NotesServer) handleSearchNotes(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: search_notes")

	query := request.GetString("query", "")
	tagsStr := request.GetString("tags", "")

	// At least one of query or tags must be provided
	if query == "" && tagsStr == "" {
		return nil, fmt.Errorf("at least one of 'query' or 'tags' parameters must be provided")
	}

	// Default limit: 1 for vector search, 10 for text search
	useVector := request.GetBool("use_vector", true)
	defaultLimit := 10
	if useVector && s.vectorSearch != nil {
		defaultLimit = 1 // Vector search defaults to top result only
	}
	limit := request.GetInt("limit", defaultLimit)

	var notes []*models.Note
	var err error

	// Handle tag search
	if tagsStr != "" {
		// Parse tags
		var tags []string
		for _, tag := range strings.Split(tagsStr, ",") {
			cleanTag := strings.TrimSpace(tag)
			if cleanTag != "" {
				tags = append(tags, cleanTag)
			}
		}
		notes, err = s.repo.SearchByTags(tags)
	} else if useVector && s.vectorSearch != nil {
		notes, err = s.vectorSearch.SearchSimilar(query, limit)
	} else {
		notes, err = s.repo.Search(query)
	}

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Limit results if text search returned too many
	if len(notes) > limit {
		notes = notes[:limit]
	}

	// Format results
	var result string
	if len(notes) == 0 {
		result = "No notes found matching your query."
	} else if len(notes) == 1 {
		// Special formatting for single result (common with vector search)
		note := notes[0]
		result = fmt.Sprintf("Most similar note:\n\n[ID: %d] %s\n\n%s",
			note.ID, note.Title,
			truncateString(note.Content, 200)) // Show more content for single result
	} else {
		result = fmt.Sprintf("Found %d notes:\n\n", len(notes))
		for i, note := range notes {
			tagsInfo := ""
			if len(note.Tags) > 0 {
				tagsInfo = fmt.Sprintf(" [Tags: %s]", strings.Join(note.Tags, ", "))
			}
			result += fmt.Sprintf("%d. [ID: %d] %s%s\n   %s\n\n",
				i+1, note.ID, note.Title, tagsInfo,
				truncateString(note.Content, 100))
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (s *NotesServer) handleGetNote(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: get_note")

	id, err := request.RequireInt("id")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'id': %w", err)
	}

	note, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	result := fmt.Sprintf("Note ID: %d\nTitle: %s", note.ID, note.Title)
	if len(note.Tags) > 0 {
		result += fmt.Sprintf("\nTags: %s", strings.Join(note.Tags, ", "))
	}
	result += fmt.Sprintf("\nCreated: %s\nUpdated: %s\n\nContent:\n%s",
		note.CreatedAt.Format("2006-01-02 15:04:05"),
		note.UpdatedAt.Format("2006-01-02 15:04:05"),
		note.Content)

	return mcp.NewToolResultText(result), nil
}

func (s *NotesServer) handleListNotes(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: list_notes")

	limit := request.GetInt("limit", 50)
	offset := request.GetInt("offset", 0)

	notes, err := s.repo.ListWithLimit(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	// Format results
	var result string
	if len(notes) == 0 {
		result = "No notes found."
	} else {
		result = fmt.Sprintf("Listing %d notes (offset: %d):\n\n", len(notes), offset)
		for i, note := range notes {
			tagsInfo := ""
			if len(note.Tags) > 0 {
				tagsInfo = fmt.Sprintf(" [Tags: %s]", strings.Join(note.Tags, ", "))
			}
			result += fmt.Sprintf("%d. [ID: %d] %s%s (Created: %s)\n   %s\n\n",
				i+1+offset, note.ID, note.Title, tagsInfo,
				note.CreatedAt.Format("2006-01-02"),
				truncateString(note.Content, 80))
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (s *NotesServer) handleUpdateNote(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: update_note")

	id, err := request.RequireInt("id")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'id': %w", err)
	}

	// Get existing note
	note, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Update fields if provided
	title := request.GetString("title", "")
	if title != "" {
		note.Title = title
	}

	content := request.GetString("content", "")
	if content != "" {
		note.Content = content
	}

	// Update tags if provided
	tagsStr := request.GetString("tags", "")
	if tagsStr != "" {
		var tags []string
		for _, tag := range strings.Split(tagsStr, ",") {
			cleanTag := strings.TrimSpace(tag)
			if cleanTag != "" {
				tags = append(tags, cleanTag)
			}
		}
		if err := s.repo.UpdateTags(note.ID, tags); err != nil {
			return nil, fmt.Errorf("failed to update tags: %w", err)
		}
	}

	// Update in database
	if err := s.repo.Update(note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// Re-index for vector search
	if s.vectorSearch != nil {
		fullText := note.Title + " " + note.Content
		if err := s.vectorSearch.IndexNote(note.ID, fullText); err != nil {
			logger.Error("Failed to re-index note %d: %v", note.ID, err)
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Note %d updated successfully.\nTitle: %s", note.ID, note.Title)), nil
}

func (s *NotesServer) handleDeleteNote(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: delete_note")

	id, err := request.RequireInt("id")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'id': %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return nil, fmt.Errorf("failed to delete note: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted note %d", id)), nil
}

// Resource handlers
func (s *NotesServer) handleRecentNotes(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	logger.Debug("MCP resource read: notes://recent")

	notes, err := s.repo.ListWithLimit(10, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent notes: %w", err)
	}

	// Format as readable text
	var content string
	content = "Recent Notes:\n\n"
	for i, note := range notes {
		content += fmt.Sprintf("%d. [ID: %d] %s\n   Created: %s\n   %s\n\n",
			i+1, note.ID, note.Title,
			note.CreatedAt.Format("2006-01-02 15:04:05"),
			truncateString(note.Content, 150))
	}

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			Text: content,
		},
	}, nil
}

func (s *NotesServer) handleStats(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	logger.Debug("MCP resource read: notes://stats")

	// Get total count
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to get note count: %w", err)
	}

	content := fmt.Sprintf(`Notes Database Statistics:
- Total Notes: %d
- Database Path: %s
- Lil-rag URL: %s`,
		count,
		s.cfg.GetDatabasePath(),
		s.cfg.LilRagURL)

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			Text: content,
		},
	}, nil
}

func (s *NotesServer) handleConfig(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	logger.Debug("MCP resource read: notes://config")

	content := fmt.Sprintf(`ML Notes Configuration:
- Debug Mode: %v
- Data Directory: %s
- Lil-rag URL: %s`,
		s.cfg.Debug,
		s.cfg.DataDirectory,
		s.cfg.LilRagURL)

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			Text: content,
		},
	}, nil
}

// Prompt handlers
func (s *NotesServer) handleSearchPrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	query := request.Params.Arguments["query"]
	limitStr := request.Params.Arguments["limit"]
	limit := 10
	if limitStr != "" {
		_, _ = fmt.Sscanf(limitStr, "%d", &limit)
	}

	prompt := fmt.Sprintf("Search for notes matching: %s\n\nPlease search for up to %d notes that match this query.", query, limit)
	return &mcp.GetPromptResult{
		Description: "Search prompt for notes",
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(prompt),
			},
		},
	}, nil
}

func (s *NotesServer) handleSummarizePrompt(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	notes, err := s.repo.List(100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	var content string
	content = "Please summarize the following notes:\n\n"
	for i, note := range notes {
		content += fmt.Sprintf("Note %d - %s:\n%s\n\n", i+1, note.Title, note.Content)
		if i >= 20 { // Limit to first 20 notes for summary
			content += fmt.Sprintf("... and %d more notes\n", len(notes)-20)
			break
		}
	}

	return &mcp.GetPromptResult{
		Description: "Summary prompt for all notes",
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(content),
			},
		},
	}, nil
}

func (s *NotesServer) handleListTags(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: list_tags")

	tags, err := s.repo.GetAllTags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var result string
	if len(tags) == 0 {
		result = "No tags found."
	} else {
		result = fmt.Sprintf("Found %d tags:\n\n", len(tags))
		for i, tag := range tags {
			result += fmt.Sprintf("%d. %s (Created: %s)\n",
				i+1, tag.Name, tag.CreatedAt.Format("2006-01-02"))
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (s *NotesServer) handleUpdateNoteTags(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: update_note_tags")

	id, err := request.RequireInt("id")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'id': %w", err)
	}

	tagsStr, err := request.RequireString("tags")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'tags': %w", err)
	}

	// Verify note exists
	_, err = s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			cleanTag := strings.TrimSpace(tag)
			if cleanTag != "" {
				tags = append(tags, cleanTag)
			}
		}
	}

	// Update tags
	if err := s.repo.UpdateTags(id, tags); err != nil {
		return nil, fmt.Errorf("failed to update tags: %w", err)
	}

	result := fmt.Sprintf("Successfully updated tags for note %d", id)
	if len(tags) > 0 {
		result += fmt.Sprintf("\nTags: %s", strings.Join(tags, ", "))
	} else {
		result += "\nRemoved all tags from note"
	}

	return mcp.NewToolResultText(result), nil
}

func (s *NotesServer) handleSuggestTags(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: suggest_tags")

	id, err := request.RequireInt("id")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'id': %w", err)
	}

	// Get the note
	note, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Check if auto-tagging is available
	if !s.autoTagger.IsAvailable() {
		return nil, fmt.Errorf("auto-tagging is not available. Please ensure auto-tagging and summarization are enabled, and Ollama is accessible")
	}

	// Get tag suggestions
	suggestedTags, err := s.autoTagger.SuggestTags(note)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tag suggestions: %w", err)
	}

	result := fmt.Sprintf("Tag suggestions for note %d (\"%s\"):", id, note.Title)
	if len(suggestedTags) > 0 {
		result += fmt.Sprintf("\n\nSuggested tags: %s", strings.Join(suggestedTags, ", "))
		if len(note.Tags) > 0 {
			result += fmt.Sprintf("\nExisting tags: %s", strings.Join(note.Tags, ", "))
		}
	} else {
		result += "\n\nNo tag suggestions generated."
	}

	return mcp.NewToolResultText(result), nil
}

func (s *NotesServer) handleAutoTagNote(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logger.Debug("MCP tool call: auto_tag_note")

	id, err := request.RequireInt("id")
	if err != nil {
		return nil, fmt.Errorf("missing required parameter 'id': %w", err)
	}

	overwrite := request.GetBool("overwrite", false)

	// Get the note
	note, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Check if auto-tagging is available
	if !s.autoTagger.IsAvailable() {
		return nil, fmt.Errorf("auto-tagging is not available. Please ensure auto-tagging and summarization are enabled, and Ollama is accessible")
	}

	// Get tag suggestions
	suggestedTags, err := s.autoTagger.SuggestTags(note)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tag suggestions: %w", err)
	}

	if len(suggestedTags) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No tags were suggested for note %d", id)), nil
	}

	// Determine final tags
	var finalTags []string
	if overwrite || len(note.Tags) == 0 {
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

	// Apply tags
	if err := s.repo.UpdateTags(id, finalTags); err != nil {
		return nil, fmt.Errorf("failed to apply tags: %w", err)
	}

	result := fmt.Sprintf("Successfully auto-tagged note %d (\"%s\")", id, note.Title)
	result += fmt.Sprintf("\nGenerated tags: %s", strings.Join(suggestedTags, ", "))
	result += fmt.Sprintf("\nApplied tags: %s", strings.Join(finalTags, ", "))

	return mcp.NewToolResultText(result), nil
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
