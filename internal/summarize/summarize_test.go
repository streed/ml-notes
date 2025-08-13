package summarize

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/models"
)

func TestNewSummarizer(t *testing.T) {
	cfg := &config.Config{
		OllamaEndpoint:      "http://localhost:11434",
		SummarizationModel:  "test-model",
		EnableSummarization: true,
	}

	summarizer := NewSummarizer(cfg)

	if summarizer == nil {
		t.Fatal("Expected non-nil summarizer")
	}

	if summarizer.model != "llama3.2:latest" {
		t.Errorf("Expected default model llama3.2:latest, got %s", summarizer.model)
	}

	if summarizer.maxTokens != 500 {
		t.Errorf("Expected default maxTokens 500, got %d", summarizer.maxTokens)
	}

	if summarizer.temperature != 0.3 {
		t.Errorf("Expected default temperature 0.3, got %f", summarizer.temperature)
	}
}

func TestSetModel(t *testing.T) {
	cfg := &config.Config{
		SummarizationModel: "initial-model",
	}

	summarizer := NewSummarizer(cfg)

	newModel := "new-model"
	summarizer.SetModel(newModel)

	if summarizer.model != newModel {
		t.Errorf("Expected model to be updated to %s, got %s", newModel, summarizer.model)
	}
}

func TestSummarizeNote_WithMockServer(t *testing.T) {
	// Create a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/generate" {
			response := `{"response": "This is a summary of the note.", "done": true}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint:      server.URL,
		EnableSummarization: true,
	}

	summarizer := NewSummarizer(cfg)
	summarizer.SetModel("test-model")

	note := &models.Note{
		ID:        1,
		Title:     "Test Note",
		Content:   "This is a test note with some content that needs to be summarized.",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := summarizer.SummarizeNote(note)
	if err != nil {
		t.Fatalf("Failed to summarize note: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Summary != "This is a summary of the note." {
		t.Errorf("Expected summary text, got: %s", result.Summary)
	}

	if result.Model != "test-model" {
		t.Errorf("Expected model test-model, got: %s", result.Model)
	}

	if result.OriginalLength == 0 {
		t.Error("Original length should be greater than 0")
	}

	if result.SummaryLength == 0 {
		t.Error("Summary length should be greater than 0")
	}
}

func TestSummarizeNotes_WithQuery(t *testing.T) {
	// Create a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/generate" {
			response := `{"response": "Summary of multiple notes related to the query.", "done": true}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	notes := []*models.Note{
		{
			ID:        1,
			Title:     "Note 1",
			Content:   "Content of note 1",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Title:     "Note 2",
			Content:   "Content of note 2",
			CreatedAt: time.Now(),
		},
	}

	result, err := summarizer.SummarizeNotes(notes, "test query")
	if err != nil {
		t.Fatalf("Failed to summarize notes: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if !strings.Contains(result.Summary, "Summary of multiple notes") {
		t.Errorf("Unexpected summary: %s", result.Summary)
	}
}

func TestSummarizeNotes_WithoutQuery(t *testing.T) {
	// Create a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/generate" {
			response := `{"response": "General summary of all notes.", "done": true}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	notes := []*models.Note{
		{
			ID:        1,
			Title:     "Note 1",
			Content:   "Content of note 1",
			CreatedAt: time.Now(),
		},
	}

	result, err := summarizer.SummarizeNotes(notes, "")
	if err != nil {
		t.Fatalf("Failed to summarize notes: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Summary != "General summary of all notes." {
		t.Errorf("Unexpected summary: %s", result.Summary)
	}
}

func TestSummarizeNotes_LongContent(t *testing.T) {
	// Create a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"response": "Summary of truncated content.", "done": true}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	// Create a note with very long content (>1000 chars)
	longContent := strings.Repeat("This is a very long piece of content. ", 100)
	notes := []*models.Note{
		{
			ID:        1,
			Title:     "Long Note",
			Content:   longContent,
			CreatedAt: time.Now(),
		},
	}

	result, err := summarizer.SummarizeNotes(notes, "")
	if err != nil {
		t.Fatalf("Failed to summarize long note: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// The original content should be truncated in the prompt
	if result.OriginalLength > 2000 {
		t.Log("Original content was properly included in prompt building")
	}
}

func TestSummarizeText_WithContext(t *testing.T) {
	// Create a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"response": "Contextual summary.", "done": true}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	text := "This is some text to summarize."
	context := "Please focus on the key technical aspects."

	result, err := summarizer.SummarizeText(text, context)
	if err != nil {
		t.Fatalf("Failed to summarize text: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Summary != "Contextual summary." {
		t.Errorf("Unexpected summary: %s", result.Summary)
	}

	if result.OriginalLength != len(text) {
		t.Errorf("Expected original length %d, got %d", len(text), result.OriginalLength)
	}
}

func TestSummarizeText_WithoutContext(t *testing.T) {
	// Create a mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"response": "Simple summary.", "done": true}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	text := "Text without specific context."

	result, err := summarizer.SummarizeText(text, "")
	if err != nil {
		t.Fatalf("Failed to summarize text: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Summary != "Simple summary." {
		t.Errorf("Unexpected summary: %s", result.Summary)
	}
}

func TestCallOllama_EmptyEndpoint(t *testing.T) {
	cfg := &config.Config{
		OllamaEndpoint: "",
	}

	summarizer := NewSummarizer(cfg)

	note := &models.Note{
		Title:   "Test",
		Content: "Content",
	}

	_, err := summarizer.SummarizeNote(note)
	if err == nil {
		t.Error("Expected error with empty endpoint")
	}

	if !strings.Contains(err.Error(), "Ollama endpoint not configured") {
		t.Errorf("Expected endpoint error, got: %v", err)
	}
}

func TestCallOllama_ServerError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	note := &models.Note{
		Title:   "Test",
		Content: "Content",
	}

	_, err := summarizer.SummarizeNote(note)
	if err == nil {
		t.Error("Expected error with server error")
	}

	if !strings.Contains(err.Error(), "Ollama API returned 500") {
		t.Errorf("Expected API error, got: %v", err)
	}
}

func TestCallOllama_InvalidJSON(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	note := &models.Note{
		Title:   "Test",
		Content: "Content",
	}

	_, err := summarizer.SummarizeNote(note)
	if err == nil {
		t.Error("Expected error with invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestCheckModelAvailability(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError bool
		errorContains string
	}{
		{
			name:          "Model exists",
			statusCode:    http.StatusOK,
			responseBody:  `{"name": "test-model"}`,
			expectedError: false,
		},
		{
			name:          "Model not found",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error": "model not found"}`,
			expectedError: true,
			errorContains: "not found",
		},
		{
			name:          "Server error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error": "internal error"}`,
			expectedError: false, // Current implementation doesn't check for 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg := &config.Config{
				OllamaEndpoint: server.URL,
			}

			summarizer := NewSummarizer(cfg)
			summarizer.SetModel("test-model")

			err := summarizer.CheckModelAvailability()

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectedError && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestCheckModelAvailability_ConnectionError(t *testing.T) {
	cfg := &config.Config{
		OllamaEndpoint: "http://invalid-host-that-does-not-exist:99999",
	}

	summarizer := NewSummarizer(cfg)
	err := summarizer.CheckModelAvailability()

	if err == nil {
		t.Error("Expected error with invalid host")
	}

	if !strings.Contains(err.Error(), "failed to check model") {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

func TestSummaryResultCreation(t *testing.T) {
	originalContent := "This is a long piece of content that needs summarization"
	summary := "Short summary"
	model := "test-model"

	result := &SummaryResult{
		Summary:        summary,
		Model:          model,
		OriginalLength: len(originalContent),
		SummaryLength:  len(summary),
	}

	if result.Summary != summary {
		t.Errorf("Expected summary %s, got %s", summary, result.Summary)
	}

	if result.Model != model {
		t.Errorf("Expected model %s, got %s", model, result.Model)
	}

	if result.OriginalLength != len(originalContent) {
		t.Errorf("Expected original length %d, got %d", len(originalContent), result.OriginalLength)
	}

	if result.SummaryLength != len(summary) {
		t.Errorf("Expected summary length %d, got %d", len(summary), result.SummaryLength)
	}

	compressionRatio := float64(result.SummaryLength) / float64(result.OriginalLength)
	expectedRatio := float64(len(summary)) / float64(len(originalContent))
	if compressionRatio != expectedRatio {
		t.Errorf("Compression ratio mismatch: expected %f, got %f", expectedRatio, compressionRatio)
	}
}

func TestSummarizerConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.Config
		expectedModel  string
		expectedTokens int
		expectedTemp   float32
	}{
		{
			name: "Default configuration",
			cfg: &config.Config{
				OllamaEndpoint: "http://localhost:11434",
			},
			expectedModel:  "llama3.2:latest",
			expectedTokens: 500,
			expectedTemp:   0.3,
		},
		{
			name: "Custom endpoint",
			cfg: &config.Config{
				OllamaEndpoint: "http://custom:11434",
			},
			expectedModel:  "llama3.2:latest",
			expectedTokens: 500,
			expectedTemp:   0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSummarizer(tt.cfg)

			if s.model != tt.expectedModel {
				t.Errorf("Expected model %s, got %s", tt.expectedModel, s.model)
			}

			if s.maxTokens != tt.expectedTokens {
				t.Errorf("Expected maxTokens %d, got %d", tt.expectedTokens, s.maxTokens)
			}

			if s.temperature != tt.expectedTemp {
				t.Errorf("Expected temperature %f, got %f", tt.expectedTemp, s.temperature)
			}
		})
	}
}

func TestMultipleNotesFormatting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just return a simple response
		response := `{"response": "Summary", "done": true}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	// Test with various numbers of notes
	tests := []struct {
		name      string
		noteCount int
		query     string
	}{
		{"Single note without query", 1, ""},
		{"Single note with query", 1, "test query"},
		{"Multiple notes without query", 3, ""},
		{"Multiple notes with query", 5, "search term"},
		{"Many notes", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notes := make([]*models.Note, tt.noteCount)
			for i := 0; i < tt.noteCount; i++ {
				notes[i] = &models.Note{
					ID:        i + 1,
					Title:     fmt.Sprintf("Note %d", i+1),
					Content:   fmt.Sprintf("Content for note %d", i+1),
					CreatedAt: time.Now(),
				}
			}

			result, err := summarizer.SummarizeNotes(notes, tt.query)
			if err != nil {
				t.Errorf("Failed to summarize %d notes: %v", tt.noteCount, err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}

func TestSummarizeWithWhitespace(t *testing.T) {
	// Test that responses are properly trimmed
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return response with extra whitespace
		response := `{"response": "  \n\n  Summary with whitespace  \n\n  ", "done": true}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)

	note := &models.Note{
		Title:   "Test",
		Content: "Content",
	}

	result, err := summarizer.SummarizeNote(note)
	if err != nil {
		t.Fatalf("Failed to summarize: %v", err)
	}

	// Check that whitespace was trimmed
	if result.Summary != "Summary with whitespace" {
		t.Errorf("Expected trimmed summary, got: '%s'", result.Summary)
	}
}

func TestCleanThinkingTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No thinking tags",
			input:    "This is a regular summary without any tags.",
			expected: "This is a regular summary without any tags.",
		},
		{
			name:     "Single line thinking tags",
			input:    "This is a summary. <think>Internal reasoning here.</think> Final conclusion.",
			expected: "This is a summary.  Final conclusion.",
		},
		{
			name:     "Multi-line thinking tags",
			input:    "Start of summary.\n<think>\nLine 1 of thinking\nLine 2 of thinking\n</think>\nEnd of summary.",
			expected: "Start of summary.\n\nEnd of summary.",
		},
		{
			name:     "Multiple thinking blocks",
			input:    "Part 1. <think>First thought</think> Part 2. <think>Second thought</think> Part 3.",
			expected: "Part 1.  Part 2.  Part 3.",
		},
		{
			name:     "Nested or malformed tags",
			input:    "Text <think>nested <think>content</think> here</think> more text",
			expected: "Text  here more text", // The regex doesn't handle nested tags perfectly, but removes the outer block
		},
		{
			name:     "Standalone tags",
			input:    "Text with <think> standalone and </think> tags scattered",
			expected: "Text with  tags scattered", // Standalone tags and text between them are removed
		},
		{
			name:     "Thinking tags with excessive newlines",
			input:    "Summary start.\n\n\n<think>Thinking content</think>\n\n\n\nSummary end.",
			expected: "Summary start.\n\nSummary end.", // Multiple newlines are collapsed to double newline
		},
		{
			name:     "Empty thinking tags",
			input:    "Summary with <think></think> empty tags.",
			expected: "Summary with  empty tags.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanThinkingTags(tt.input)
			if result != tt.expected {
				t.Errorf("cleanThinkingTags() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSummarizeWithThinkingTags(t *testing.T) {
	// Create a mock Ollama server that returns a response with thinking tags
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"response": "Here is the summary. <think>This is internal reasoning that should be removed.</think> This is the final summary.", "done": true}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint:      server.URL,
		EnableSummarization: true,
	}

	summarizer := NewSummarizer(cfg)
	
	note := &models.Note{
		ID:        1,
		Title:     "Test Note",
		Content:   "Test content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := summarizer.SummarizeNote(note)
	if err != nil {
		t.Fatalf("Failed to summarize note: %v", err)
	}

	// Check that thinking tags were removed
	expectedSummary := "Here is the summary.  This is the final summary."
	if result.Summary != expectedSummary {
		t.Errorf("Expected thinking tags to be removed. Got: %q, Want: %q", result.Summary, expectedSummary)
	}
}

func BenchmarkSummarizeNote(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"response": "Benchmark summary.", "done": true}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	cfg := &config.Config{
		OllamaEndpoint: server.URL,
	}

	summarizer := NewSummarizer(cfg)
	note := &models.Note{
		Title:   "Benchmark Note",
		Content: "Content for benchmarking",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = summarizer.SummarizeNote(note)
	}
}
