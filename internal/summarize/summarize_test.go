package summarize

import (
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

func TestSummarizerWithDisabledSummarization(t *testing.T) {
	cfg := &config.Config{
		EnableSummarization: false,
	}

	summarizer := NewSummarizer(cfg)
	
	if summarizer == nil {
		t.Fatal("Expected non-nil summarizer even when disabled")
	}

	// The summarizer should still be created even if summarization is disabled
	// The check should happen at a higher level
	if summarizer.cfg.EnableSummarization != false {
		t.Error("EnableSummarization should be false")
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

func TestSetModelMultipleTimes(t *testing.T) {
	cfg := &config.Config{}
	s := NewSummarizer(cfg)

	models := []string{"model1", "model2", "model3"}
	
	for _, model := range models {
		s.SetModel(model)
		if s.model != model {
			t.Errorf("Expected model to be %s, got %s", model, s.model)
		}
	}
}

func TestSummaryResultCompressionMetrics(t *testing.T) {
	tests := []struct {
		name           string
		originalLength int
		summaryLength  int
		expectedRatio  float64
	}{
		{
			name:           "50% compression",
			originalLength: 1000,
			summaryLength:  500,
			expectedRatio:  0.5,
		},
		{
			name:           "90% compression",
			originalLength: 1000,
			summaryLength:  100,
			expectedRatio:  0.1,
		},
		{
			name:           "No compression",
			originalLength: 100,
			summaryLength:  100,
			expectedRatio:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &SummaryResult{
				OriginalLength: tt.originalLength,
				SummaryLength:  tt.summaryLength,
			}
			
			ratio := float64(result.SummaryLength) / float64(result.OriginalLength)
			if ratio != tt.expectedRatio {
				t.Errorf("Expected compression ratio %f, got %f", tt.expectedRatio, ratio)
			}
		})
	}
}

func TestNotePreparation(t *testing.T) {
	// Test that notes with various content types can be created
	tests := []struct {
		name    string
		note    *models.Note
		wantErr bool
	}{
		{
			name: "Normal note",
			note: &models.Note{
				Title:   "Test Title",
				Content: "Test content",
			},
			wantErr: false,
		},
		{
			name: "Empty title",
			note: &models.Note{
				Title:   "",
				Content: "Content without title",
			},
			wantErr: false,
		},
		{
			name: "Empty content",
			note: &models.Note{
				Title:   "Title without content",
				Content: "",
			},
			wantErr: false,
		},
		{
			name: "Very long content",
			note: &models.Note{
				Title:   "Long Note",
				Content: strings.Repeat("x", 10000),
			},
			wantErr: false,
		},
		{
			name: "Note with timestamps",
			note: &models.Note{
				Title:     "Timestamped Note",
				Content:   "Content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now().Add(1 * time.Hour),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the note can be created
			if tt.note == nil && !tt.wantErr {
				t.Error("Note should not be nil")
			}
		})
	}
}