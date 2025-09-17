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

The MCP server provides comprehensive tools and resources that enable LLMs to:

Tools (10 available):
- add_note: Create new notes with optional tags
- search_notes: Enhanced search with vector/text/tag modes and flexible output
- get_note: Retrieve specific notes by ID
- list_notes: List notes with pagination
- update_note: Modify existing notes and tags
- delete_note: Remove notes from database
- list_tags: View all available tags
- update_note_tags: Manage note tags
- suggest_tags: AI-powered tag suggestions
- auto_tag_note: Automatically apply AI-generated tags

Resources (6 available):
- notes://recent: Most recently created notes
- notes://note/{id}: Individual note access by ID
- notes://tags: Complete tag listing with metadata
- notes://stats: Comprehensive database statistics
- notes://config: System configuration and capabilities
- notes://health: Service health and availability status

Prompts (2 available):
- search_notes: Structured search interactions
- summarize_notes: Generate analysis of note collections

To use with Claude Desktop, add this to your claude_desktop_config.json:
{
  "mcpServers": {
    "ml-notes": {
      "command": "ml-notes-cli",
      "args": ["mcp"]
    }
  }
}

For CLI binary usage, ensure you're using 'ml-notes-cli' as the command.`,
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
