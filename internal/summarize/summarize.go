package summarize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
)

// Summarizer provides text summarization capabilities
type Summarizer struct {
	cfg         *config.Config
	model       string
	maxTokens   int
	temperature float32
}

// SummaryResult contains the summarized content
type SummaryResult struct {
	Summary        string
	OriginalLength int
	SummaryLength  int
	Model          string
}

// NewSummarizer creates a new summarizer instance
func NewSummarizer(cfg *config.Config) *Summarizer {
	return &Summarizer{
		cfg:         cfg,
		model:       "llama3.2:latest", // Default model for summarization
		maxTokens:   500,
		temperature: 0.3, // Lower temperature for more focused summaries
	}
}

// SetModel allows changing the model used for summarization
func (s *Summarizer) SetModel(model string) {
	s.model = model
}

// SummarizeNote creates a summary of a single note
func (s *Summarizer) SummarizeNote(note *models.Note) (*SummaryResult, error) {
	content := fmt.Sprintf("Title: %s\n\nContent:\n%s", note.Title, note.Content)

	prompt := fmt.Sprintf(`Please provide a concise summary of the following note. 
Focus on the key points and main ideas. Keep the summary brief but informative.

%s

Summary:`, content)

	summary, err := s.callOllama(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	return &SummaryResult{
		Summary:        summary,
		OriginalLength: len(content),
		SummaryLength:  len(summary),
		Model:          s.model,
	}, nil
}

// SummarizeNotes creates a combined summary of multiple notes
func (s *Summarizer) SummarizeNotes(notes []*models.Note, query string) (*SummaryResult, error) {
	var contentBuilder strings.Builder

	if query != "" {
		contentBuilder.WriteString(fmt.Sprintf("Search Query: '%s'\n\n", query))
		contentBuilder.WriteString("Search Results:\n\n")
	}

	for i, note := range notes {
		contentBuilder.WriteString(fmt.Sprintf("Note %d (ID: %d)\n", i+1, note.ID))
		contentBuilder.WriteString(fmt.Sprintf("Title: %s\n", note.Title))
		contentBuilder.WriteString(fmt.Sprintf("Created: %s\n", note.CreatedAt.Format("2006-01-02")))

		// Truncate very long notes in multi-note summaries
		content := note.Content
		if len(content) > 1000 {
			content = content[:1000] + "..."
		}
		contentBuilder.WriteString(fmt.Sprintf("Content: %s\n\n", content))
		contentBuilder.WriteString("---\n\n")
	}

	fullContent := contentBuilder.String()

	var prompt string
	if query != "" {
		prompt = fmt.Sprintf(`Please provide a comprehensive summary of the following search results.
The user searched for: "%s"

Analyze how these notes relate to the search query and highlight the most relevant information.
Group related topics together and identify common themes.

%s

Summary:`, query, fullContent)
	} else {
		prompt = fmt.Sprintf(`Please provide a comprehensive summary of the following notes.
Identify key themes, patterns, and important information across all notes.
Group related topics together and highlight the most significant points.

%s

Summary:`, fullContent)
	}

	summary, err := s.callOllama(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	return &SummaryResult{
		Summary:        summary,
		OriginalLength: len(fullContent),
		SummaryLength:  len(summary),
		Model:          s.model,
	}, nil
}

// SummarizeText creates a summary of arbitrary text
func (s *Summarizer) SummarizeText(text string, context string) (*SummaryResult, error) {
	var prompt string
	if context != "" {
		prompt = fmt.Sprintf(`%s

Text to summarize:
%s

Summary:`, context, text)
	} else {
		prompt = fmt.Sprintf(`Please provide a concise summary of the following text.
Focus on the key points and main ideas.

%s

Summary:`, text)
	}

	summary, err := s.callOllama(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	return &SummaryResult{
		Summary:        summary,
		OriginalLength: len(text),
		SummaryLength:  len(summary),
		Model:          s.model,
	}, nil
}

// callOllama makes a request to the Ollama API for text generation
func (s *Summarizer) callOllama(prompt string) (string, error) {
	if s.cfg.OllamaEndpoint == "" {
		return "", fmt.Errorf("Ollama endpoint not configured")
	}

	payload := map[string]interface{}{
		"model":       s.model,
		"prompt":      prompt,
		"temperature": s.temperature,
		"stream":      false,
		"options": map[string]interface{}{
			"num_predict": s.maxTokens,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := s.cfg.GetOllamaAPIURL("generate")
	logger.Debug("Requesting summary from %s with model %s", apiURL, s.model)

	start := time.Now()
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Ollama API error: %v", err)
		return "", fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug("Ollama response status: %d, time: %v", resp.StatusCode, time.Since(start))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return strings.TrimSpace(result.Response), nil
}

// CheckModelAvailability verifies if the summarization model is available
func (s *Summarizer) CheckModelAvailability() error {
	payload := map[string]interface{}{
		"name": s.model,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := s.cfg.GetOllamaAPIURL("show")
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to check model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("model %s not found. Please pull it first with: ollama pull %s", s.model, s.model)
	}

	return nil
}
