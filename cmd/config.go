package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/constants"
	interrors "github.com/streed/ml-notes/internal/errors"
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
  - debug: Enable/disable debug logging (true/false)
  - summarization-model: Model to use for summarization
  - enable-summarization: Enable/disable summarization features (true/false)
  - editor: Default editor to use for editing notes (e.g., "vim", "code --wait")`,
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
	fmt.Printf("Config file:           %s\n", configPath)
	fmt.Printf("data-dir:              %s\n", cfg.DataDirectory)
	fmt.Printf("Database path:         %s\n", cfg.GetDatabasePath())
	fmt.Printf("ollama-endpoint:       %s\n", cfg.OllamaEndpoint)
	fmt.Printf("embedding-model:       %s\n", cfg.EmbeddingModel)
	fmt.Printf("vector-dimensions:     %d\n", cfg.VectorDimensions)
	fmt.Printf("enable-vector:         %v\n", cfg.EnableVectorSearch)
	fmt.Printf("debug:                 %v\n", cfg.Debug)
	fmt.Printf("enable-summarization:  %v\n", cfg.EnableSummarization)
	if cfg.EnableSummarization {
		fmt.Printf("summarization-model:   %s\n", cfg.SummarizationModel)
	}
	if cfg.Editor != "" {
		fmt.Printf("editor:                %s\n", cfg.Editor)
	} else {
		fmt.Printf("editor:                Auto-detect\n")
	}
	fmt.Println("SQLite-vec:            Built-in (via Go bindings)")
	if cfg.VectorConfigVersion != "" {
		fmt.Printf("Vector config hash:    %s\n", cfg.VectorConfigVersion)
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
			return fmt.Errorf("%w: %s", interrors.ErrInvalidDimensions, value)
		}
		if cfg.VectorDimensions != dims {
			needsReindex = true
		}
		cfg.VectorDimensions = dims
	case "enable-vector":
		var enable bool
		if value == constants.BoolTrue || value == constants.BoolOne || value == constants.BoolYes {
			enable = true
		} else if value == constants.BoolFalse || value == constants.BoolZero || value == constants.BoolNo {
			enable = false
		} else {
			return fmt.Errorf("%w: %s", interrors.ErrInvalidBoolean, value)
		}
		if cfg.EnableVectorSearch != enable {
			needsReindex = true
		}
		cfg.EnableVectorSearch = enable
	case "debug":
		var debug bool
		if value == constants.BoolTrue || value == constants.BoolOne || value == constants.BoolYes {
			debug = true
		} else if value == constants.BoolFalse || value == constants.BoolZero || value == constants.BoolNo {
			debug = false
		} else {
			return fmt.Errorf("%w: %s", interrors.ErrInvalidBoolean, value)
		}
		cfg.Debug = debug
	case "summarization-model":
		cfg.SummarizationModel = value
	case "enable-summarization":
		var enable bool
		if value == constants.BoolTrue || value == constants.BoolOne || value == constants.BoolYes {
			enable = true
		} else if value == constants.BoolFalse || value == constants.BoolZero || value == constants.BoolNo {
			enable = false
		} else {
			return fmt.Errorf("%w: %s", interrors.ErrInvalidBoolean, value)
		}
		cfg.EnableSummarization = enable
	case "editor":
		cfg.Editor = value
	default:
		return fmt.Errorf("%w: %s", interrors.ErrUnknownConfigKey, key)
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
