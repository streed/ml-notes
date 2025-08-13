package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ml-notes configuration",
	Long: `Initialize ml-notes configuration interactively or with flags.
This command sets up the configuration file and creates necessary directories.`,
	RunE: runInit,
}

var (
	initDataDir             string
	initOllamaEndpoint      string
	initInteractive         bool
	initSummarizationModel  string
	initEnableSummarization bool
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initDataDir, "data-dir", "", "Data directory for storing notes database")
	initCmd.Flags().StringVar(&initOllamaEndpoint, "ollama-endpoint", "", "Ollama API endpoint (e.g., http://localhost:11434)")
	initCmd.Flags().StringVar(&initSummarizationModel, "summarization-model", "", "Model to use for summarization (e.g., llama3.2:latest)")
	initCmd.Flags().BoolVar(&initEnableSummarization, "enable-summarization", true, "Enable AI summarization features")
	initCmd.Flags().BoolVarP(&initInteractive, "interactive", "i", false, "Run interactive setup")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration already exists at: %s\n", configPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Configuration initialization cancelled.")
			return nil
		}
	}

	// Set defaults for non-interactive mode if not provided
	if !initInteractive {
		if initSummarizationModel == "" {
			initSummarizationModel = "llama3.2:latest"
		}
	}

	// Interactive mode
	if initInteractive || (initDataDir == "" && initOllamaEndpoint == "") {
		fmt.Println("=== ML Notes Configuration Setup ===")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)

		// Data directory
		defaultDataDir := config.GetDefaultDataDirectory()
		fmt.Printf("Data directory [%s]: ", defaultDataDir)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			initDataDir = expandPath(input)
		} else {
			initDataDir = defaultDataDir
		}

		// Ollama endpoint
		fmt.Printf("Ollama API endpoint [http://localhost:11434]: ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			initOllamaEndpoint = input
		} else {
			initOllamaEndpoint = "http://localhost:11434"
		}

		// Summarization settings
		fmt.Println("\n--- Summarization Settings ---")
		fmt.Printf("Enable AI summarization features? (Y/n): ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "n" || input == "no" {
			initEnableSummarization = false
		} else {
			initEnableSummarization = true

			// Summarization model
			fmt.Printf("Summarization model [llama3.2:latest]: ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				initSummarizationModel = input
			} else {
				initSummarizationModel = "llama3.2:latest"
			}
		}
	}

	// Create configuration
	cfg, err := config.InitializeConfigWithSummarization(
		initDataDir,
		initOllamaEndpoint,
		initSummarizationModel,
		initEnableSummarization,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Display summary
	fmt.Println("\n=== Configuration Summary ===")
	fmt.Printf("Config file:        %s\n", configPath)
	fmt.Printf("Data directory:     %s\n", cfg.DataDirectory)
	fmt.Printf("Database path:      %s\n", cfg.GetDatabasePath())
	fmt.Printf("Ollama endpoint:    %s\n", cfg.OllamaEndpoint)
	fmt.Printf("Embedding model:    %s\n", cfg.EmbeddingModel)
	fmt.Printf("Vector dimensions:  %d\n", cfg.VectorDimensions)
	fmt.Printf("Vector search:      %v\n", cfg.EnableVectorSearch)
	fmt.Printf("Summarization:      %v\n", cfg.EnableSummarization)
	if cfg.EnableSummarization {
		fmt.Printf("Summarize model:    %s\n", cfg.SummarizationModel)
	}
	fmt.Println("SQLite-vec:         Built-in (via Go bindings)")

	fmt.Println("\nConfiguration initialized successfully!")
	fmt.Println("You can now use 'ml-notes' commands to manage your notes.")

	// Test Ollama connection
	fmt.Print("\nWould you like to test the Ollama connection? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "y" || response == "yes" {
		testOllamaConnection(cfg)
	}

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

func testOllamaConnection(cfg *config.Config) {
	fmt.Printf("\nTesting connection to Ollama at %s...\n", cfg.OllamaEndpoint)

	// We'll implement a simple ping to the Ollama API
	// For now, just show a message
	fmt.Println("Note: Make sure Ollama is running and has the required models installed:")
	fmt.Printf("  ollama pull %s  # For embeddings\n", cfg.EmbeddingModel)
	if cfg.EnableSummarization && cfg.SummarizationModel != "" {
		fmt.Printf("  ollama pull %s  # For summarization\n", cfg.SummarizationModel)
	}
}
