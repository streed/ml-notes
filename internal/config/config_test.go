package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultDataDirectory(t *testing.T) {
	tests := []struct {
		name     string
		xdgHome  string
		expected string
	}{
		{
			name:     "With XDG_DATA_HOME set",
			xdgHome:  "/custom/data",
			expected: "/custom/data/ml-notes",
		},
		{
			name:     "Without XDG_DATA_HOME",
			xdgHome:  "",
			expected: "ml-notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldXDG := os.Getenv("XDG_DATA_HOME")
			defer os.Setenv("XDG_DATA_HOME", oldXDG)

			os.Setenv("XDG_DATA_HOME", tt.xdgHome)
			result := GetDefaultDataDirectory()

			if tt.xdgHome == "" {
				homeDir, _ := os.UserHomeDir()
				expected := filepath.Join(homeDir, ".local", "share", "ml-notes")
				if result != expected {
					t.Errorf("Expected %s, got %s", expected, result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "ml-notes", "config.json")

	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	// Use temp directory for data directory to avoid permission issues
	dataDir := filepath.Join(tempDir, "test-data")
	
	testConfig := &Config{
		DataDirectory:       dataDir,
		DatabasePath:        filepath.Join(dataDir, "notes.db"),
		OllamaEndpoint:      "http://test:11434",
		EmbeddingModel:      "test-model",
		VectorDimensions:    768,
		EnableVectorSearch:  true,
		Debug:               true,
		SummarizationModel:  "test-summary-model",
		EnableSummarization: true,
	}

	// Test Save
	err := Save(testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Test Load
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Compare configs
	if loadedConfig.DataDirectory != testConfig.DataDirectory {
		t.Errorf("DataDirectory mismatch: expected %s, got %s",
			testConfig.DataDirectory, loadedConfig.DataDirectory)
	}
	if loadedConfig.OllamaEndpoint != testConfig.OllamaEndpoint {
		t.Errorf("OllamaEndpoint mismatch: expected %s, got %s",
			testConfig.OllamaEndpoint, loadedConfig.OllamaEndpoint)
	}
	if loadedConfig.EmbeddingModel != testConfig.EmbeddingModel {
		t.Errorf("EmbeddingModel mismatch: expected %s, got %s",
			testConfig.EmbeddingModel, loadedConfig.EmbeddingModel)
	}
	if loadedConfig.VectorDimensions != testConfig.VectorDimensions {
		t.Errorf("VectorDimensions mismatch: expected %d, got %d",
			testConfig.VectorDimensions, loadedConfig.VectorDimensions)
	}
	if loadedConfig.EnableVectorSearch != testConfig.EnableVectorSearch {
		t.Errorf("EnableVectorSearch mismatch: expected %v, got %v",
			testConfig.EnableVectorSearch, loadedConfig.EnableVectorSearch)
	}
	if loadedConfig.Debug != testConfig.Debug {
		t.Errorf("Debug mismatch: expected %v, got %v",
			testConfig.Debug, loadedConfig.Debug)
	}
	if loadedConfig.SummarizationModel != testConfig.SummarizationModel {
		t.Errorf("SummarizationModel mismatch: expected %s, got %s",
			testConfig.SummarizationModel, loadedConfig.SummarizationModel)
	}
	if loadedConfig.EnableSummarization != testConfig.EnableSummarization {
		t.Errorf("EnableSummarization mismatch: expected %v, got %v",
			testConfig.EnableSummarization, loadedConfig.EnableSummarization)
	}
}

func TestInitializeConfig(t *testing.T) {
	tempDir := t.TempDir()
	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	dataDir := filepath.Join(tempDir, "data")
	ollamaEndpoint := "http://custom:11434"

	cfg, err := InitializeConfig(dataDir, ollamaEndpoint)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	if cfg.DataDirectory != dataDir {
		t.Errorf("Expected DataDirectory %s, got %s", dataDir, cfg.DataDirectory)
	}

	expectedDBPath := filepath.Join(dataDir, "notes.db")
	if cfg.DatabasePath != expectedDBPath {
		t.Errorf("Expected DatabasePath %s, got %s", expectedDBPath, cfg.DatabasePath)
	}

	if cfg.OllamaEndpoint != ollamaEndpoint {
		t.Errorf("Expected OllamaEndpoint %s, got %s", ollamaEndpoint, cfg.OllamaEndpoint)
	}

	// Check that config file was created
	configFile := filepath.Join(tempDir, "ml-notes", "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created during initialization")
	}
}

func TestInitializeConfigWithSummarization(t *testing.T) {
	tempDir := t.TempDir()
	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	dataDir := filepath.Join(tempDir, "data")
	ollamaEndpoint := "http://custom:11434"
	summarizationModel := "llama3:latest"
	enableSummarization := true

	cfg, err := InitializeConfigWithSummarization(dataDir, ollamaEndpoint, summarizationModel, enableSummarization)
	if err != nil {
		t.Fatalf("Failed to initialize config with summarization: %v", err)
	}

	if cfg.SummarizationModel != summarizationModel {
		t.Errorf("Expected SummarizationModel %s, got %s", summarizationModel, cfg.SummarizationModel)
	}

	if cfg.EnableSummarization != enableSummarization {
		t.Errorf("Expected EnableSummarization %v, got %v", enableSummarization, cfg.EnableSummarization)
	}
}

func TestGetDatabasePath(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		expectedPath string
	}{
		{
			name: "With DatabasePath set",
			config: Config{
				DatabasePath:  "/custom/path/notes.db",
				DataDirectory: "/data",
			},
			expectedPath: "/custom/path/notes.db",
		},
		{
			name: "Without DatabasePath set",
			config: Config{
				DatabasePath:  "",
				DataDirectory: "/data",
			},
			expectedPath: "/data/notes.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDatabasePath()
			if result != tt.expectedPath {
				t.Errorf("Expected %s, got %s", tt.expectedPath, result)
			}
		})
	}
}

func TestGetOllamaAPIURL(t *testing.T) {
	cfg := Config{
		OllamaEndpoint: "http://localhost:11434",
	}

	tests := []struct {
		endpoint string
		expected string
	}{
		{"embeddings", "http://localhost:11434/api/embeddings"},
		{"generate", "http://localhost:11434/api/generate"},
		{"tags", "http://localhost:11434/api/tags"},
	}

	for _, tt := range tests {
		result := cfg.GetOllamaAPIURL(tt.endpoint)
		if result != tt.expected {
			t.Errorf("For endpoint %s: expected %s, got %s", tt.endpoint, tt.expected, result)
		}
	}
}

func TestGetVectorConfigHash(t *testing.T) {
	cfg := Config{
		EmbeddingModel:     "test-model",
		VectorDimensions:   384,
		EnableVectorSearch: true,
	}

	hash := cfg.GetVectorConfigHash()
	expected := "test-model-384-true"

	if hash != expected {
		t.Errorf("Expected hash %s, got %s", expected, hash)
	}
}

func TestNeedsReindex(t *testing.T) {
	cfg := Config{
		EmbeddingModel:     "model-v1",
		VectorDimensions:   384,
		EnableVectorSearch: true,
	}

	oldHash := "model-v1-384-true"
	if cfg.NeedsReindex(oldHash) {
		t.Error("Should not need reindex with same hash")
	}

	differentHash := "model-v2-768-true"
	if !cfg.NeedsReindex(differentHash) {
		t.Error("Should need reindex with different hash")
	}
}

func TestLoadWithDefaults(t *testing.T) {
	tempDir := t.TempDir()
	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	// Create a partial config file
	configDir := filepath.Join(tempDir, "ml-notes")
	os.MkdirAll(configDir, 0755)

	partialConfig := map[string]interface{}{
		"embedding_model": "custom-model",
		// Intentionally leave out other fields to test defaults
	}

	data, _ := json.MarshalIndent(partialConfig, "", "  ")
	configFile := filepath.Join(configDir, "config.json")
	os.WriteFile(configFile, data, 0600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check that custom value was loaded
	if cfg.EmbeddingModel != "custom-model" {
		t.Errorf("Expected custom EmbeddingModel, got %s", cfg.EmbeddingModel)
	}

	// Check that defaults were applied (OllamaEndpoint should be set to default since we didn't specify it)
	if cfg.OllamaEndpoint != "http://localhost:11434" {
		t.Errorf("Expected default OllamaEndpoint 'http://localhost:11434', got '%s'", cfg.OllamaEndpoint)
	}

	if cfg.VectorDimensions != 384 {
		t.Errorf("Expected default VectorDimensions 384, got %d", cfg.VectorDimensions)
	}
}