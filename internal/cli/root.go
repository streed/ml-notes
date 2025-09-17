package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/api"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/database"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/search"
)

var (
	db            *database.DB
	noteRepo      *models.NoteRepository
	vectorSearch  search.SearchProvider
	appConfig     *config.Config
	assetProvider api.AssetProvider
	debugFlag     bool
	Version       = "dev" // Version is set from main.go
)

var rootCmd = &cobra.Command{
	Use:     "ml-notes-cli",
	Short:   "A CLI tool for managing notes with semantic search",
	Version: Version,
	Long: `ml-notes-cli is a command-line interface for creating, managing, and searching notes using lil-rag for semantic search.

First time users should run 'ml-notes-cli init' to set up the configuration.

Note: For a desktop GUI experience, use the main 'ml-notes' executable.`,
}

func Execute() error {
	rootCmd.Version = Version
	return rootCmd.Execute()
}

// SetAssetProvider sets the asset provider for embedded web assets
func SetAssetProvider(provider api.AssetProvider) {
	assetProvider = provider
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
		fmt.Fprintf(os.Stderr, "Please run 'ml-notes-cli init' to set up the configuration.\n")
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
	}

	db, err = database.New(appConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	noteRepo = models.NewNoteRepository(db.Conn())

	// Use lil-rag for search and indexing
	vectorSearch = search.NewLilRagSearch(noteRepo, appConfig)
	logger.Debug("Using lil-rag search at: %s", appConfig.LilRagURL)
}

// getCurrentProjectNamespace returns the current project namespace based on working directory
func getCurrentProjectNamespace() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}
