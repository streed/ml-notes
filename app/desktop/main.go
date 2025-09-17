package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/database"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/preferences"
	"github.com/streed/ml-notes/internal/search"
	"github.com/streed/ml-notes/internal/services"
)

// App struct
type App struct {
	ctx      context.Context
	services *services.Services
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// OnStartup is called when the app starts up.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// Initialize configuration with default values if loading fails
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		// Create a basic config with default data directory
		dataDir := "/home/reed/.local/share/ml-notes" // Default fallback
		ollamaEndpoint := "http://localhost:11434"    // Default Ollama endpoint
		cfg, err = config.InitializeConfig(dataDir, ollamaEndpoint)
		if err != nil {
			logger.Error("Failed to initialize default config: %v", err)
			return
		}
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		return
	}

	// Initialize repositories
	noteRepo := models.NewNoteRepository(db.Conn())
	prefsRepo := preferences.NewPreferencesRepository(db.Conn())

	// Initialize search (optional)
	var vectorSearch search.SearchProvider
	// vectorSearch = search.NewSQLiteVectorSearch(db.Conn()) // if available

	// Initialize services layer
	a.services = services.NewServices(cfg, noteRepo, prefsRepo, vectorSearch)
}

// OnDomReady is called after front-end resources have been loaded
func (a *App) OnDomReady(ctx context.Context) {
	// Called when the frontend is ready
}

// OnBeforeClose is called when the application is about to quit
func (a *App) OnBeforeClose(ctx context.Context) (prevent bool) {
	// Return true to prevent the application from quitting
	return false
}

// OnShutdown is called when the application is shutting down
func (a *App) OnShutdown(ctx context.Context) {
	// Cleanup resources
	if a.services != nil {
		if err := a.services.Close(); err != nil {
			logger.Error("Failed to close services: %v", err)
		}
	}
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ML Notes",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.OnStartup,
		OnDomReady:       app.OnDomReady,
		OnBeforeClose:    app.OnBeforeClose,
		OnShutdown:       app.OnShutdown,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
