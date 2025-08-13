package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ml-notes configuration",
	Long:  `View and manage ml-notes configuration settings.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current ml-notes configuration settings.`,
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  `Display the path to the configuration file.`,
	RunE:  runConfigPath,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set a specific configuration value.

Available keys:
  - data-dir: Data directory for storing notes database
  - ollama-endpoint: Ollama API endpoint
  - embedding-model: Embedding model name
  - vector-dimensions: Number of vector dimensions
  - enable-vector: Enable/disable vector search (true/false)
  - debug: Enable/disable debug logging (true/false)`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	fmt.Println("=== ML Notes Configuration ===")
	fmt.Printf("Config file:        %s\n", configPath)
	fmt.Printf("Data directory:     %s\n", cfg.DataDirectory)
	fmt.Printf("Database path:      %s\n", cfg.GetDatabasePath())
	fmt.Printf("Ollama endpoint:    %s\n", cfg.OllamaEndpoint)
	fmt.Printf("Embedding model:    %s\n", cfg.EmbeddingModel)
	fmt.Printf("Vector dimensions:  %d\n", cfg.VectorDimensions)
	fmt.Printf("Vector search:      %v\n", cfg.EnableVectorSearch)
	fmt.Printf("Debug mode:         %v\n", cfg.Debug)
	fmt.Println("SQLite-vec:         Built-in (via Go bindings)")
	if cfg.VectorConfigVersion != "" {
		fmt.Printf("Vector config hash: %s\n", cfg.VectorConfigVersion)
	}

	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	fmt.Println(configPath)
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	oldVectorHash := cfg.GetVectorConfigHash()
	needsReindex := false

	switch key {
	case "data-dir":
		cfg.DataDirectory = expandPath(value)
		cfg.DatabasePath = "" // Will be regenerated
	case "ollama-endpoint":
		cfg.OllamaEndpoint = value
	case "embedding-model":
		if cfg.EmbeddingModel != value {
			needsReindex = true
		}
		cfg.EmbeddingModel = value
	case "vector-dimensions":
		var dims int
		if _, err := fmt.Sscanf(value, "%d", &dims); err != nil {
			return fmt.Errorf("invalid vector dimensions: %s", value)
		}
		if cfg.VectorDimensions != dims {
			needsReindex = true
		}
		cfg.VectorDimensions = dims
	case "enable-vector":
		var enable bool
		if value == "true" || value == "1" || value == "yes" {
			enable = true
		} else if value == "false" || value == "0" || value == "no" {
			enable = false
		} else {
			return fmt.Errorf("invalid boolean value: %s (use true/false)", value)
		}
		if cfg.EnableVectorSearch != enable {
			needsReindex = true
		}
		cfg.EnableVectorSearch = enable
	case "debug":
		var debug bool
		if value == "true" || value == "1" || value == "yes" {
			debug = true
		} else if value == "false" || value == "0" || value == "no" {
			debug = false
		} else {
			return fmt.Errorf("invalid boolean value: %s (use true/false)", value)
		}
		cfg.Debug = debug
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Check if vector configuration changed
	if needsReindex && oldVectorHash != cfg.GetVectorConfigHash() {
		fmt.Println("\nWarning: Vector configuration has changed.")
		fmt.Println("You should run 'ml-notes reindex' to update all note embeddings.")
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}

func runConfigJSON(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode configuration: %w", err)
	}

	return nil
}