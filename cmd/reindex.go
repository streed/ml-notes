package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/embeddings"
	"github.com/streed/ml-notes/internal/logger"
)

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Reindex all notes with current vector configuration",
	Long: `Reindex all notes using the current embedding model and vector dimensions.
This is necessary after changing the embedding model or vector dimensions.`,
	RunE: runReindex,
}

var forceReindex bool

func init() {
	rootCmd.AddCommand(reindexCmd)
	reindexCmd.Flags().BoolVarP(&forceReindex, "force", "f", false, "Force reindex even if configuration hasn't changed")
}

func runReindex(cmd *cobra.Command, args []string) error {
	// Check if reindexing is needed
	currentHash := appConfig.GetVectorConfigHash()
	if !forceReindex && appConfig.VectorConfigVersion == currentHash {
		fmt.Println("Vector configuration hasn't changed. Use --force to reindex anyway.")
		return nil
	}

	if !appConfig.EnableVectorSearch {
		return fmt.Errorf("vector search is disabled in configuration")
	}

	fmt.Printf("Reindexing notes with:\n")
	fmt.Printf("  Model: %s\n", appConfig.EmbeddingModel)
	fmt.Printf("  Dimensions: %d\n", appConfig.VectorDimensions)
	fmt.Println()

	// First, detect actual dimensions from the model
	fmt.Println("Detecting model dimensions...")
	embedder := embeddings.NewLocalEmbedding(appConfig)
	// Use document type for dimension detection
	testEmbedding, err := embedder.GetEmbeddingWithType("test", embeddings.EmbeddingTypeDocument)
	if err != nil && testEmbedding == nil {
		// Configuration was auto-updated
		fmt.Println("Configuration was automatically updated with correct dimensions.")
		fmt.Println("Please run this command again to continue reindexing.")
		return nil
	}

	if testEmbedding != nil {
		actualDims := len(testEmbedding)
		if actualDims != appConfig.VectorDimensions {
			logger.Info("Model returns %d dimensions, updating configuration", actualDims)
			appConfig.VectorDimensions = actualDims
			if err := config.Save(appConfig); err != nil {
				return fmt.Errorf("failed to update configuration: %w", err)
			}
			fmt.Printf("Updated configuration to %d dimensions\n\n", actualDims)
		}
	}

	// First, recreate the vector table with new dimensions
	logger.Debug("Dropping old vector table...")
	if err := recreateVectorTable(); err != nil {
		return fmt.Errorf("failed to recreate vector table: %w", err)
	}

	// Get all notes
	notes, err := noteRepo.List(0, 0) // Get all notes
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes to reindex.")
		return nil
	}

	fmt.Printf("Reindexing %d notes...\n", len(notes))

	// Reindex each note
	successCount := 0
	for i, note := range notes {
		logger.Debug("Reindexing note %d: %s", note.ID, note.Title)

		// Create combined text for embedding
		fullText := note.Title + " " + note.Content

		// Index the note
		if err := vectorSearch.IndexNote(note.ID, fullText); err != nil {
			logger.Error("Failed to reindex note %d: %v", note.ID, err)
			fmt.Fprintf(os.Stderr, "Failed to reindex note %d (%s): %v\n", note.ID, note.Title, err)
		} else {
			successCount++
			if (i+1)%10 == 0 || i == len(notes)-1 {
				fmt.Printf("Progress: %d/%d notes\n", i+1, len(notes))
			}
		}
	}

	fmt.Printf("\nReindexing complete: %d/%d notes successfully reindexed.\n", successCount, len(notes))

	// Update the vector config version
	appConfig.VectorConfigVersion = currentHash
	if err := config.Save(appConfig); err != nil {
		logger.Error("Failed to update configuration: %v", err)
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

func recreateVectorTable() error {
	// Drop the old vector table
	_, err := db.Conn().Exec("DROP TABLE IF EXISTS vec_notes")
	if err != nil {
		logger.Debug("Failed to drop vec_notes table: %v", err)
	}

	// Clear the embeddings table
	_, err = db.Conn().Exec("DELETE FROM note_embeddings")
	if err != nil {
		logger.Debug("Failed to clear note_embeddings: %v", err)
	}

	// Recreate the vector table with new dimensions
	dimensions := appConfig.VectorDimensions
	if dimensions == 0 {
		dimensions = 384
	}

	query := fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS vec_notes USING vec0(
			note_id INTEGER PRIMARY KEY,
			embedding float[%d]
		)
	`, dimensions)

	_, err = db.Conn().Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create vec_notes table: %w", err)
	}

	logger.Debug("Recreated vec_notes table with %d dimensions", dimensions)
	return nil
}
