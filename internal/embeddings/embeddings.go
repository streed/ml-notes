package embeddings

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
)

type EmbeddingType string

const (
	EmbeddingTypeDocument EmbeddingType = "document"
	EmbeddingTypeSearch   EmbeddingType = "search"
)

type EmbeddingProvider interface {
	GetEmbedding(text string) ([]float32, error)
	GetEmbeddingWithType(text string, embedType EmbeddingType) ([]float32, error)
}

type LocalEmbedding struct {
	cfg *config.Config
}

func NewLocalEmbedding(cfg *config.Config) *LocalEmbedding {
	return &LocalEmbedding{cfg: cfg}
}

// formatTextForNomic formats text according to Nomic's recommendations
// See: https://docs.nomic.ai/reference/endpoints/nomic-embed-text
func (e *LocalEmbedding) formatTextForNomic(text string, embedType EmbeddingType) string {
	// Check if we're using a Nomic model
	if !strings.Contains(strings.ToLower(e.cfg.EmbeddingModel), "nomic") {
		return text
	}
	
	// Format based on embedding type
	switch embedType {
	case EmbeddingTypeSearch:
		// For search queries, use "search_query: " prefix
		return "search_query: " + text
	case EmbeddingTypeDocument:
		// For documents, use "search_document: " prefix
		return "search_document: " + text
	default:
		return text
	}
}

func (e *LocalEmbedding) GetEmbedding(text string) ([]float32, error) {
	// Default to document type for backward compatibility
	return e.GetEmbeddingWithType(text, EmbeddingTypeDocument)
}

func (e *LocalEmbedding) GetEmbeddingWithType(text string, embedType EmbeddingType) ([]float32, error) {
	if !e.cfg.EnableVectorSearch {
		logger.Debug("Vector search disabled, using fallback embedding")
		return e.fallbackEmbedding(text), nil
	}

	// Format text for Nomic models
	formattedText := e.formatTextForNomic(text, embedType)
	
	// Use Ollama API for embeddings
	payload := map[string]interface{}{
		"model":  e.cfg.EmbeddingModel,
		"prompt": formattedText,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := e.cfg.GetOllamaAPIURL("embeddings")
	logger.Debug("Requesting %s embedding from %s with model %s", embedType, apiURL, e.cfg.EmbeddingModel)
	if formattedText != text {
		logger.Debug("Formatted text with Nomic prefix for %s", embedType)
	}
	
	start := time.Now()
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Debug("Ollama API error: %v, using fallback embedding", err)
		return e.fallbackEmbedding(text), nil
	}
	defer resp.Body.Close()

	logger.Debug("Ollama response status: %d, time: %v", resp.StatusCode, time.Since(start))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Debug("Ollama API returned %d: %s, using fallback embedding", resp.StatusCode, string(body))
		return e.fallbackEmbedding(text), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debug("Failed to read response: %v, using fallback embedding", err)
		return e.fallbackEmbedding(text), nil
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		logger.Debug("Failed to parse embedding response: %v, using fallback embedding", err)
		return e.fallbackEmbedding(text), nil
	}

	actualDimensions := len(result.Embedding)
	logger.Debug("Got embedding with %d dimensions", actualDimensions)
	
	// Check if dimensions match configuration
	if e.cfg.VectorDimensions > 0 && actualDimensions != e.cfg.VectorDimensions {
		logger.Error("Dimension mismatch: model returned %d dimensions but config expects %d", actualDimensions, e.cfg.VectorDimensions)
		logger.Info("Updating configuration to match model dimensions: %d", actualDimensions)
		
		// Update the configuration with actual dimensions
		e.cfg.VectorDimensions = actualDimensions
		if err := config.Save(e.cfg); err != nil {
			logger.Error("Failed to update configuration: %v", err)
			return nil, fmt.Errorf("dimension mismatch: model returns %d dimensions but config expects %d. Failed to update config: %w", 
				actualDimensions, e.cfg.VectorDimensions, err)
		}
		
		logger.Info("Configuration updated. Please restart the application or run 'ml-notes reindex' to rebuild the vector table with %d dimensions", actualDimensions)
		return nil, fmt.Errorf("configuration updated to %d dimensions. Please restart or run 'ml-notes reindex'", actualDimensions)
	}
	
	return result.Embedding, nil
}


func (e *LocalEmbedding) fallbackEmbedding(text string) []float32 {
	// Simple hash-based embedding for demo purposes
	// In production, you'd want to use a real embedding model
	dimensions := e.cfg.VectorDimensions
	if dimensions == 0 {
		dimensions = 384
	}
	embedding := make([]float32, dimensions)
	words := strings.Fields(strings.ToLower(text))
	
	for i, word := range words {
		hash := hashString(word)
		for j := 0; j < dimensions && j < len(word)*10; j++ {
			idx := (i*10 + j) % dimensions
			embedding[idx] += float32(hash%100) / 100.0
		}
	}
	
	// Normalize
	var sum float32
	for _, v := range embedding {
		sum += v * v
	}
	if sum > 0 {
		norm := float32(1.0 / float32(sum))
		for i := range embedding {
			embedding[i] *= norm
		}
	}
	
	return embedding
}

func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}

func EmbeddingToBytes(embedding []float32) []byte {
	buf := new(bytes.Buffer)
	for _, v := range embedding {
		binary.Write(buf, binary.LittleEndian, v)
	}
	return buf.Bytes()
}

func BytesToEmbedding(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("invalid embedding data length")
	}
	
	embedding := make([]float32, len(data)/4)
	buf := bytes.NewReader(data)
	for i := range embedding {
		if err := binary.Read(buf, binary.LittleEndian, &embedding[i]); err != nil {
			return nil, err
		}
	}
	return embedding, nil
}

func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	
	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dotProduct / (float32(normA) * float32(normB))
}