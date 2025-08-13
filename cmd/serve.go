package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/api"
	"github.com/streed/ml-notes/internal/logger"
)

var (
	serveHost string
	servePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP API server",
	Long: `Start an HTTP API server that exposes ml-notes functionality via REST endpoints.

This allows integration with tools like Ollama, OpenWebUI, and other applications
that can consume HTTP APIs. The server provides endpoints for:

- Notes CRUD operations
- Search (text and vector)
- Tag management  
- Auto-tagging with AI
- Statistics and configuration

The API is documented at http://host:port/api/v1/docs when the server is running.

Examples:
  ml-notes serve                          # Start on localhost:8080
  ml-notes serve --host 0.0.0.0 --port 3000  # Start on all interfaces, port 3000
  
Integration with OpenWebUI:
  You can configure OpenWebUI to use this API by adding custom functions
  that call the ml-notes endpoints for note management and search.

Integration with Ollama:
  Use the API endpoints in your Ollama model configurations or tools
  to enable note-taking and retrieval capabilities.`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&serveHost, "host", "localhost", "Host to bind the server to")
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to bind the server to")
}

func runServe(cmd *cobra.Command, args []string) error {
	logger.Info("Initializing HTTP API server...")

	// Create API server
	apiServer := api.NewAPIServer(appConfig, db.Conn(), noteRepo, vectorSearch)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- apiServer.Start(serveHost, servePort)
	}()

	// Print useful information
	fmt.Printf("\n🚀 ML Notes HTTP API Server\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📍 Server URL: http://%s:%d\n", serveHost, servePort)
	fmt.Printf("📖 API Docs:   http://%s:%d/api/v1/docs\n", serveHost, servePort)
	fmt.Printf("🔍 Health:     http://%s:%d/api/v1/health\n", serveHost, servePort)
	fmt.Printf("📊 Stats:      http://%s:%d/api/v1/stats\n", serveHost, servePort)
	fmt.Printf("\n🎯 Example API calls:\n")
	fmt.Printf("   curl http://%s:%d/api/v1/notes\n", serveHost, servePort)
	fmt.Printf("   curl http://%s:%d/api/v1/tags\n", serveHost, servePort)
	fmt.Printf("   curl http://%s:%d/api/v1/health\n", serveHost, servePort)
	fmt.Printf("\n💡 For OpenWebUI integration:\n")
	fmt.Printf("   - Create custom functions that call these endpoints\n")
	fmt.Printf("   - Use the search endpoint for RAG functionality\n")
	fmt.Printf("   - Use auto-tagging for intelligent note organization\n")
	fmt.Printf("\n✋ Press Ctrl+C to stop the server\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		logger.Info("Received signal %v, shutting down gracefully...", sig)
		if err := apiServer.Stop(); err != nil {
			logger.Error("Error during server shutdown: %v", err)
			return err
		}
		logger.Info("Server stopped successfully")
		return nil
	case err := <-errChan:
		if err != nil {
			logger.Error("Server error: %v", err)
			return err
		}
		return nil
	}
}