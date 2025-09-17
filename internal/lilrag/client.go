package lilrag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	cfg        *config.Config
}

type IndexRequest struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Namespace string `json:"namespace,omitempty"`
}

type IndexResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type SearchRequest struct {
	Query     string `json:"query"`
	Limit     int    `json:"limit,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type SearchResult struct {
	ID       string                 `json:"ID"`
	Text     string                 `json:"Text"`
	Score    float64                `json:"Score"`
	Metadata map[string]interface{} `json:"Metadata,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

func NewClient(cfg *config.Config) *Client {
	baseURL := cfg.LilRagURL
	if baseURL == "" {
		baseURL = "http://localhost:8081" // Default lil-rag port
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cfg: cfg,
	}
}

func (c *Client) IndexDocument(id, text string) error {
	return c.IndexDocumentWithNamespace(id, text, "")
}

func (c *Client) IndexDocumentWithNamespace(id, text, namespace string) error {
	req := IndexRequest{
		ID:        id,
		Text:      text,
		Namespace: namespace,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal index request: %w", err)
	}

	url := c.baseURL + "/api/index"
	if namespace != "" {
		logger.Debug("Indexing document %s to lil-rag at %s (namespace: %s)", id, url, namespace)
	} else {
		logger.Debug("Indexing document %s to lil-rag at %s", id, url)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send index request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debug("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("lil-rag index request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var indexResp IndexResponse
	if err := json.NewDecoder(resp.Body).Decode(&indexResp); err != nil {
		return fmt.Errorf("failed to decode index response: %w", err)
	}

	// Check for success using either Success field (new format) or Status field (actual lil-rag format)
	success := indexResp.Success || indexResp.Status == "indexed"
	if !success {
		message := indexResp.Message
		if message == "" && indexResp.Status != "" {
			message = indexResp.Status
		}
		return fmt.Errorf("lil-rag index failed: %s", message)
	}

	message := indexResp.Message
	if message == "" && indexResp.Status != "" {
		message = indexResp.Status
	}
	logger.Debug("Successfully indexed document %s: %s", indexResp.ID, message)
	return nil
}

func (c *Client) Search(query string, limit int) ([]SearchResult, error) {
	return c.SearchWithNamespace(query, limit, "")
}

func (c *Client) SearchWithNamespace(query string, limit int, namespace string) ([]SearchResult, error) {
	req := SearchRequest{
		Query:     query,
		Limit:     limit,
		Namespace: namespace,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	url := c.baseURL + "/api/search"
	if namespace != "" {
		logger.Debug("Searching lil-rag for: %s (limit: %d, namespace: %s)", query, limit, namespace)
	} else {
		logger.Debug("Searching lil-rag for: %s (limit: %d)", query, limit)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send search request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debug("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lil-rag search request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	logger.Debug("Found %d results from lil-rag", len(searchResp.Results))
	return searchResp.Results, nil
}

func (c *Client) IsAvailable() bool {
	// Try a simple GET request to the API root or health endpoint
	endpoints := []string{"/api/search", "/health", "/"}

	for _, endpoint := range endpoints {
		url := c.baseURL + endpoint
		resp, err := c.httpClient.Get(url)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()

		// Accept any response that's not a connection error
		if resp.StatusCode < 500 {
			logger.Debug("Lil-rag service is available at %s (status: %d)", url, resp.StatusCode)
			return true
		}
	}

	logger.Debug("Lil-rag service not available at %s", c.baseURL)
	return false
}
