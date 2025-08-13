package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DatabasePath       string `json:"database_path"`
	DataDirectory      string `json:"data_directory"`
	OllamaEndpoint     string `json:"ollama_endpoint"`
	EmbeddingModel     string `json:"embedding_model"`
	VectorDimensions   int    `json:"vector_dimensions"`
	EnableVectorSearch bool   `json:"enable_vector_search"`
	Debug              bool   `json:"debug"`
	VectorConfigVersion string `json:"vector_config_version,omitempty"`
}

var (
	defaultConfig = Config{
		DatabasePath:      "", // Will be set to DataDirectory/notes.db
		DataDirectory:     "", // Will be set to ~/.local/share/ml-notes
		OllamaEndpoint:    "http://localhost:11434",
		EmbeddingModel:    "nomic-embed-text",
		VectorDimensions:  384,
		EnableVectorSearch: true,
		Debug:             false,
	}
)

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	return filepath.Join(configDir, "ml-notes", "config.json"), nil
}

func GetDefaultDataDirectory() string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", ".ml-notes")
		}
		dataDir = filepath.Join(homeDir, ".local", "share")
	}
	return filepath.Join(dataDir, "ml-notes")
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		cfg := defaultConfig
		if cfg.DataDirectory == "" {
			cfg.DataDirectory = GetDefaultDataDirectory()
		}
		if cfg.DatabasePath == "" {
			cfg.DatabasePath = filepath.Join(cfg.DataDirectory, "notes.db")
		}
		return &cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for empty fields
	if cfg.DataDirectory == "" {
		cfg.DataDirectory = GetDefaultDataDirectory()
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = filepath.Join(cfg.DataDirectory, "notes.db")
	}
	if cfg.OllamaEndpoint == "" {
		cfg.OllamaEndpoint = defaultConfig.OllamaEndpoint
	}
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = defaultConfig.EmbeddingModel
	}
	if cfg.VectorDimensions == 0 {
		cfg.VectorDimensions = defaultConfig.VectorDimensions
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create data directory if it doesn't exist
	if cfg.DataDirectory != "" {
		if err := os.MkdirAll(cfg.DataDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func InitializeConfig(dataDir, ollamaEndpoint string) (*Config, error) {
	cfg := &defaultConfig

	// Set custom values if provided
	if dataDir != "" {
		cfg.DataDirectory = dataDir
	} else {
		cfg.DataDirectory = GetDefaultDataDirectory()
	}
	
	cfg.DatabasePath = filepath.Join(cfg.DataDirectory, "notes.db")
	
	if ollamaEndpoint != "" {
		cfg.OllamaEndpoint = ollamaEndpoint
	}

	// Save the configuration
	if err := Save(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) GetDatabasePath() string {
	if c.DatabasePath != "" {
		return c.DatabasePath
	}
	return filepath.Join(c.DataDirectory, "notes.db")
}

func (c *Config) GetOllamaAPIURL(endpoint string) string {
	return fmt.Sprintf("%s/api/%s", c.OllamaEndpoint, endpoint)
}

func (c *Config) GetVectorConfigHash() string {
	// Create a hash of vector-related configuration
	return fmt.Sprintf("%s-%d-%v", c.EmbeddingModel, c.VectorDimensions, c.EnableVectorSearch)
}

func (c *Config) NeedsReindex(oldHash string) bool {
	return c.GetVectorConfigHash() != oldHash
}