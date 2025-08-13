package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/database"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/search"
)

var (
	db           *database.DB
	noteRepo     *models.NoteRepository
	vectorSearch *search.VectorSearch
	appConfig    *config.Config
	debugFlag    bool
	Version      = "dev" // Version is set from main.go
)

var rootCmd = &cobra.Command{
	Use:     "ml-notes",
	Short:   "A CLI tool for managing notes with vector search",
	Version: Version,
	Long: `ml-notes is a command-line interface for creating, managing, and searching notes using vector embeddings for semantic search.

First time users should run 'ml-notes init' to set up the configuration.`,
}

func Execute() error {
	rootCmd.Version = Version
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initAppConfig)
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
}

func initAppConfig() {
	// Skip initialization for init and config commands
	if len(os.Args) > 1 && (os.Args[1] == "init" || os.Args[1] == "config") {
		return
	}

	var err error
	appConfig, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please run 'ml-notes init' to set up the configuration.\n")
		os.Exit(1)
	}

	// Enable debug mode from flag or config
	if debugFlag || appConfig.Debug {
		logger.SetDebugMode(true)
		logger.Debug("Configuration loaded from: %s", func() string {
			path, _ := config.GetConfigPath()
			return path
		}())
		logger.Debug("Data directory: %s", appConfig.DataDirectory)
		logger.Debug("Ollama endpoint: %s", appConfig.OllamaEndpoint)
		logger.Debug("Vector search enabled: %v", appConfig.EnableVectorSearch)
		logger.Debug("Embedding model: %s", appConfig.EmbeddingModel)
		logger.Debug("Vector dimensions: %d", appConfig.VectorDimensions)
	}

	db, err = database.New(appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	noteRepo = models.NewNoteRepository(db.Conn())
	vectorSearch = search.NewVectorSearch(db.Conn(), noteRepo, appConfig)

	// Check if reindexing is needed
	checkAndReindex()
}

func checkAndReindex() {
	if appConfig.VectorConfigVersion != "" && appConfig.NeedsReindex(appConfig.VectorConfigVersion) {
		logger.Info("Vector configuration has changed. Reindexing is recommended.")
		logger.Info("Run 'ml-notes reindex' to update all embeddings.")
	}
}
