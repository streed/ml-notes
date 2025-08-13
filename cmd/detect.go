package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/embeddings"
	"github.com/streed/ml-notes/internal/logger"
)

var detectCmd = &cobra.Command{
	Use:   "detect-dimensions",
	Short: "Detect the actual embedding dimensions from the model",
	Long:  `Query the configured embedding model to detect its actual output dimensions.`,
	RunE:  runDetect,
}

func init() {
	rootCmd.AddCommand(detectCmd)
}

func runDetect(cmd *cobra.Command, args []string) error {
	if !appConfig.EnableVectorSearch {
		return fmt.Errorf("vector search is disabled in configuration")
	}

	fmt.Printf("Detecting dimensions for model: %s\n", appConfig.EmbeddingModel)
	fmt.Printf("Current configured dimensions: %d\n", appConfig.VectorDimensions)
	fmt.Printf("Ollama endpoint: %s\n\n", appConfig.OllamaEndpoint)

	// Create embedder
	embedder := embeddings.NewLocalEmbedding(appConfig)
	
	// Get a test embedding using document type
	testText := "This is a test to detect embedding dimensions"
	logger.Debug("Getting test embedding for: %s", testText)
	
	// Use document type for dimension detection
	embedding, err := embedder.GetEmbeddingWithType(testText, embeddings.EmbeddingTypeDocument)
	if err != nil {
		// Check if it's a dimension mismatch error
		if embedding == nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nThe configuration has been updated automatically.")
			fmt.Println("Please run 'ml-notes reindex' to rebuild the vector table.")
			return nil
		}
		return fmt.Errorf("failed to get embedding: %w", err)
	}

	actualDimensions := len(embedding)
	fmt.Printf("✓ Model returns %d dimensions\n", actualDimensions)

	if actualDimensions != appConfig.VectorDimensions {
		fmt.Printf("\nWarning: Configuration mismatch!\n")
		fmt.Printf("  Config expects: %d dimensions\n", appConfig.VectorDimensions)
		fmt.Printf("  Model returns:  %d dimensions\n", actualDimensions)
		fmt.Println("\nRun 'ml-notes config set vector-dimensions " + fmt.Sprint(actualDimensions) + "' to update the configuration")
		fmt.Println("Then run 'ml-notes reindex' to rebuild all embeddings")
	} else {
		fmt.Println("✓ Configuration matches model output")
	}

	return nil
}