package autotag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
)

type AutoTagger struct {
	cfg        *config.Config
	httpClient *http.Client
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type TagSuggestion struct {
	Tags       []string `json:"tags"`
	Confidence string   `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
}

func NewAutoTagger(cfg *config.Config) *AutoTagger {
	return &AutoTagger{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SuggestTags analyzes note content and suggests relevant tags
func (at *AutoTagger) SuggestTags(note *models.Note) ([]string, error) {
	if !at.cfg.EnableAutoTagging || !at.cfg.EnableSummarization {
		return nil, fmt.Errorf("auto-tagging requires both auto-tagging and summarization to be enabled")
	}

	prompt := at.buildTaggingPrompt(note)
	
	logger.Debug("Sending auto-tagging request to Ollama")
	response, err := at.callOllama(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag suggestions from Ollama: %w", err)
	}

	tags, err := at.parseTags(response)
	if err != nil {
		logger.Debug("Failed to parse structured response, falling back to simple extraction: %v", err)
		// Fallback to simple tag extraction
		tags = at.extractTagsFromText(response)
	}

	// Clean and validate tags
	cleanTags := at.cleanTags(tags)
	
	logger.Debug("Suggested tags for note %d: %v", note.ID, cleanTags)
	return cleanTags, nil
}

// SuggestTagsBatch processes multiple notes for auto-tagging
func (at *AutoTagger) SuggestTagsBatch(notes []*models.Note) (map[int][]string, error) {
	results := make(map[int][]string)
	
	for i, note := range notes {
		logger.Debug("Auto-tagging note %d/%d (ID: %d)", i+1, len(notes), note.ID)
		
		tags, err := at.SuggestTags(note)
		if err != nil {
			logger.Error("Failed to auto-tag note %d: %v", note.ID, err)
			// Continue with other notes
			continue
		}
		
		results[note.ID] = tags
		
		// Add a small delay to avoid overwhelming Ollama
		time.Sleep(500 * time.Millisecond)
	}
	
	return results, nil
}

func (at *AutoTagger) buildTaggingPrompt(note *models.Note) string {
	maxTags := at.cfg.MaxAutoTags
	if maxTags == 0 {
		maxTags = 5
	}
	return fmt.Sprintf(`Please analyze the following note and suggest up to %d relevant tags that would help categorize and organize it.

Title: %s
Content: %s

Instructions:
- Generate tags that capture the main topics, themes, and categories
- Use lowercase, single words or short phrases connected with hyphens
- Focus on actionable, searchable keywords
- Consider: subject matter, project type, priority level, status, context
- Return ONLY a comma-separated list of tags, no explanations

Examples of good tags: research, meeting, urgent, project-alpha, learning, todo, completed, python, ai, personal

Tags:`, maxTags, note.Title, note.Content)
}

func (at *AutoTagger) callOllama(prompt string) (string, error) {
	// Use auto-tag model if specified, otherwise use summarization model
	model := at.cfg.AutoTagModel
	if model == "" {
		model = at.cfg.SummarizationModel
	}
	
	reqBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.3, // Lower temperature for more consistent tagging
			"top_p":       0.9,
			"max_tokens":  100, // Tags should be short
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := at.httpClient.Post(at.cfg.OllamaEndpoint+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// parseTags attempts to parse structured tag response (JSON or formatted)
func (at *AutoTagger) parseTags(response string) ([]string, error) {
	// Try to find JSON in the response
	jsonRegex := regexp.MustCompile(`\{[^}]*"tags"[^}]*\}`)
	if match := jsonRegex.FindString(response); match != "" {
		var suggestion TagSuggestion
		if err := json.Unmarshal([]byte(match), &suggestion); err == nil {
			return suggestion.Tags, nil
		}
	}

	// Try to find tags after "Tags:" or similar patterns
	patterns := []string{
		`(?i)tags?:\s*(.+)`,
		`(?i)suggested tags?:\s*(.+)`,
		`(?i)categories?:\s*(.+)`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindStringSubmatch(response); len(matches) > 1 {
			return at.parseTagString(matches[1]), nil
		}
	}

	return nil, fmt.Errorf("could not parse structured tags from response")
}

// extractTagsFromText extracts tags from free-form text response
func (at *AutoTagger) extractTagsFromText(response string) []string {
	// Remove common prefixes and clean the response
	response = strings.TrimSpace(response)
	
	// Remove common response patterns
	patterns := []string{
		`(?i)^(here are|suggested|recommended)?\s*(tags?|keywords?):?\s*`,
		`(?i)^(the\s+)?(following\s+)?tags?\s+(are|would\s+be):?\s*`,
	}
	
	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		response = regex.ReplaceAllString(response, "")
	}

	return at.parseTagString(response)
}

// parseTagString parses a comma-separated or space-separated string of tags
func (at *AutoTagger) parseTagString(tagStr string) []string {
	tagStr = strings.TrimSpace(tagStr)
	
	// Handle different separators
	var tags []string
	if strings.Contains(tagStr, ",") {
		tags = strings.Split(tagStr, ",")
	} else if strings.Contains(tagStr, ";") {
		tags = strings.Split(tagStr, ";")
	} else {
		// Split by whitespace, but preserve hyphenated tags
		tags = regexp.MustCompile(`\s+`).Split(tagStr, -1)
	}

	var result []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result
}

// cleanTags validates, normalizes and deduplicates tags
func (at *AutoTagger) cleanTags(tags []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		// Basic cleanup
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		
		// Remove quotes, bullets, numbers, and other unwanted characters
		tag = regexp.MustCompile(`^[-•*\d.\s"'\[\]()]+`).ReplaceAllString(tag, "")
		tag = regexp.MustCompile(`[-•*\d.\s"'\[\]()]+$`).ReplaceAllString(tag, "")
		
		// Skip empty or invalid tags
		if tag == "" || len(tag) < 2 || len(tag) > 30 {
			continue
		}
		
		// Skip common stop words that aren't useful as tags
		stopWords := []string{
			"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with",
			"by", "from", "up", "about", "into", "through", "during", "before",
			"after", "above", "below", "between", "among", "through", "during",
			"note", "notes", "content", "text", "information", "data", "item",
		}
		
		isStopWord := false
		for _, stopWord := range stopWords {
			if tag == stopWord {
				isStopWord = true
				break
			}
		}
		
		if isStopWord {
			continue
		}
		
		// Deduplicate
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	// Limit to maximum number of tags from config
	maxTags := at.cfg.MaxAutoTags
	if maxTags == 0 {
		maxTags = 5
	}
	if len(result) > maxTags {
		result = result[:maxTags]
	}

	return result
}

// IsAvailable checks if auto-tagging is available (Ollama accessible and summarization enabled)
func (at *AutoTagger) IsAvailable() bool {
	if !at.cfg.EnableAutoTagging || !at.cfg.EnableSummarization {
		return false
	}

	// Quick health check to Ollama
	resp, err := at.httpClient.Get(at.cfg.OllamaEndpoint + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}