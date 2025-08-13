package cmd

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for LLM integration",
	Long: `Start a Model Context Protocol (MCP) server that allows LLMs to interact with your notes.
	
The MCP server provides tools and resources that enable LLMs to:
- Search notes using vector similarity or text search
- Add, update, and delete notes
- List and retrieve specific notes
- Access notes statistics and configuration

To use with Claude Desktop, add this to your claude_desktop_config.json:
{
  "mcpServers": {
    "ml-notes": {
      "command": "ml-notes",
      "args": ["mcp"]
    }
  }
}`,
	RunE: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) error {
	logger.Info("Starting MCP server...")
	
	// Create MCP server
	notesServer := mcp.NewNotesServer(appConfig, db.Conn(), noteRepo, vectorSearch)
	mcpServer := notesServer.GetMCPServer()
	
	// Start server with stdio transport
	logger.Info("MCP server ready. Listening on stdio...")
	if err := server.ServeStdio(mcpServer); err != nil {
		if err.Error() != "EOF" {
			logger.Error("MCP server error: %v", err)
			return err
		}
	}
	
	logger.Info("MCP server shutting down")
	return nil
}