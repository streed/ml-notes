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
  - debug: Enable/disable debug logging (true/false)
  - summarization-model: Model to use for summarization
  - enable-summarization: Enable/disable summarization features (true/false)
  - editor: Default editor to use for editing notes (e.g., "vim", "code --wait")
  - enable-auto-tagging: Enable/disable AI auto-tagging features (true/false)
  - auto-tag-model: Model to use for auto-tagging (leave empty to use summarization model)
  - max-auto-tags: Maximum number of tags to auto-generate per note (1-20)
  - github-owner: GitHub repository owner for updates (default: streed)
  - github-repo: GitHub repository name for updates (default: ml-notes)
  - lilrag-url: Lil-Rag service endpoint for enhanced semantic search`,
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
	fmt.Printf("enable-auto-tagging:   %v\n", cfg.EnableAutoTagging)
	if cfg.AutoTagModel != "" {
		fmt.Printf("auto-tag-model:        %s\n", cfg.AutoTagModel)
	} else {
		fmt.Printf("auto-tag-model:        %s (using summarization model)\n", cfg.SummarizationModel)
	}
	fmt.Printf("max-auto-tags:         %d\n", cfg.MaxAutoTags)
	fmt.Printf("github-owner:          %s\n", cfg.GitHubOwner)
	fmt.Printf("github-repo:           %s\n", cfg.GitHubRepo)
	fmt.Printf("lilrag-url:            %s\n", cfg.LilRagURL)

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

	switch key {
	case "data-dir":
		cfg.DataDirectory = expandPath(value)
		cfg.DatabasePath = "" // Will be regenerated
	case "ollama-endpoint":
		cfg.OllamaEndpoint = value
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
	case "enable-auto-tagging":
		var enable bool
		if value == constants.BoolTrue || value == constants.BoolOne || value == constants.BoolYes {
			enable = true
		} else if value == constants.BoolFalse || value == constants.BoolZero || value == constants.BoolNo {
			enable = false
		} else {
			return fmt.Errorf("%w: %s", interrors.ErrInvalidBoolean, value)
		}
		cfg.EnableAutoTagging = enable
	case "auto-tag-model":
		cfg.AutoTagModel = value
	case "max-auto-tags":
		var maxTags int
		if _, err := fmt.Sscanf(value, "%d", &maxTags); err != nil || maxTags < 1 || maxTags > 20 {
			return fmt.Errorf("invalid max-auto-tags value: must be between 1 and 20")
		}
		cfg.MaxAutoTags = maxTags
	case "github-owner":
		cfg.GitHubOwner = value
	case "github-repo":
		cfg.GitHubRepo = value
	case "lilrag-url":
		cfg.LilRagURL = value
	default:
		return fmt.Errorf("%w: %s", interrors.ErrUnknownConfigKey, key)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}
