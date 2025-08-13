package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	interrors "github.com/streed/ml-notes/internal/errors"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/summarize"
)

var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a note by ID",
	Long:  `Display the full content of a note by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var getSummarize bool

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVar(&getSummarize, "summarize", false, "Generate a summary of the note")
}

func runGet(_ *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("%w: %s", interrors.ErrInvalidNoteID, args[0])
	}

	note, err := noteRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	fmt.Printf("================================================================================\n")
	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", note.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("================================================================================\n\n")

	// Generate summary if requested
	if getSummarize && appConfig.EnableSummarization {
		fmt.Println("📝 Summary:")
		fmt.Println(strings.Repeat("-", 80))

		summarizer := summarize.NewSummarizer(appConfig)
		if appConfig.SummarizationModel != "" {
			summarizer.SetModel(appConfig.SummarizationModel)
		}

		result, err := summarizer.SummarizeNote(note)
		if err != nil {
			logger.Error("Failed to generate summary: %v", err)
			fmt.Printf("Warning: Could not generate summary: %v\n", err)
		} else {
			fmt.Println(result.Summary)
			fmt.Println(strings.Repeat("-", 80))
			fmt.Printf("✨ Summary generated using %s\n", result.Model)
			fmt.Printf("   Reduced from %d to %d characters (%.1f%% compression)\n",
				result.OriginalLength, result.SummaryLength,
				100.0*(1.0-float64(result.SummaryLength)/float64(result.OriginalLength)))
		}

		fmt.Printf("\n================================================================================\n")
		fmt.Println("Full Content:")
	}

	fmt.Println(note.Content)
	fmt.Println()

	return nil
}
