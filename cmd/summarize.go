package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/summarize"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize [ids...]",
	Short: "Summarize notes",
	Long: `Generate AI-powered summaries of one or more notes.
	
You can summarize:
- A single note by ID: ml-notes summarize 123
- Multiple notes by IDs: ml-notes summarize 1 2 3
- All notes: ml-notes summarize --all
- Recent notes: ml-notes summarize --recent 10`,
	RunE: runSummarize,
}

var (
	summarizeAll    bool
	summarizeRecent int
	summarizeModel  string
)

func init() {
	rootCmd.AddCommand(summarizeCmd)
	summarizeCmd.Flags().BoolVar(&summarizeAll, "all", false, "Summarize all notes")
	summarizeCmd.Flags().IntVar(&summarizeRecent, "recent", 0, "Summarize N most recent notes")
	summarizeCmd.Flags().StringVar(&summarizeModel, "model", "", "Override the summarization model")
}

func runSummarize(_ *cobra.Command, args []string) error {
	if !appConfig.EnableSummarization {
		return fmt.Errorf("summarization is disabled in configuration")
	}

	var notes []*models.Note
	var err error

	// Determine which notes to summarize
	if summarizeAll {
		fmt.Println("Loading all notes...")
		notes, err = noteRepo.List(0, 0)
		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}
	} else if summarizeRecent > 0 {
		fmt.Printf("Loading %d most recent notes...\n", summarizeRecent)
		notes, err = noteRepo.List(summarizeRecent, 0)
		if err != nil {
			return fmt.Errorf("failed to list recent notes: %w", err)
		}
	} else if len(args) > 0 {
		// Summarize specific notes by ID
		for _, arg := range args {
			id, err := strconv.Atoi(arg)
			if err != nil {
				logger.Error("Invalid note ID: %s", arg)
				continue
			}

			note, err := noteRepo.GetByID(id)
			if err != nil {
				logger.Error("Failed to get note %d: %v", id, err)
				continue
			}

			notes = append(notes, note)
		}

		if len(notes) == 0 {
			return fmt.Errorf("no valid notes found")
		}
	} else {
		return fmt.Errorf("please specify note IDs or use --all or --recent flags")
	}

	fmt.Printf("Summarizing %d note(s)...\n\n", len(notes))
	fmt.Println(strings.Repeat("=", 80))

	// Create summarizer
	summarizer := summarize.NewSummarizer(appConfig)
	if summarizeModel != "" {
		summarizer.SetModel(summarizeModel)
		fmt.Printf("Using model: %s\n\n", summarizeModel)
	} else if appConfig.SummarizationModel != "" {
		summarizer.SetModel(appConfig.SummarizationModel)
		fmt.Printf("Using model: %s\n\n", appConfig.SummarizationModel)
	}

	// Generate summary based on number of notes
	if len(notes) == 1 {
		// Single note summary
		note := notes[0]
		fmt.Printf("Note: %s (ID: %d)\n", note.Title, note.ID)
		fmt.Println(strings.Repeat("-", 80))

		result, err := summarizer.SummarizeNote(note)
		if err != nil {
			return fmt.Errorf("failed to generate summary: %w", err)
		}

		fmt.Println("\n📝 Summary:")
		fmt.Println(result.Summary)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("\n✨ Summary generated using %s\n", result.Model)
		fmt.Printf("   Reduced from %d to %d characters (%.1f%% compression)\n",
			result.OriginalLength, result.SummaryLength,
			100.0*(1.0-float64(result.SummaryLength)/float64(result.OriginalLength)))
	} else {
		// Multiple notes summary
		fmt.Printf("Summarizing %d notes together...\n", len(notes))
		fmt.Println(strings.Repeat("-", 80))

		// Show titles of notes being summarized
		fmt.Println("Notes included:")
		for i, note := range notes {
			fmt.Printf("  %d. [ID: %d] %s\n", i+1, note.ID, note.Title)
			if i >= 10 && len(notes) > 12 {
				fmt.Printf("  ... and %d more notes\n", len(notes)-11)
				break
			}
		}
		fmt.Println()

		result, err := summarizer.SummarizeNotes(notes, "")
		if err != nil {
			return fmt.Errorf("failed to generate summary: %w", err)
		}

		fmt.Println("📝 Combined Summary:")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println(result.Summary)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("\n✨ Summary generated using %s\n", result.Model)
		fmt.Printf("   Reduced from %d to %d characters (%.1f%% compression)\n",
			result.OriginalLength, result.SummaryLength,
			100.0*(1.0-float64(result.SummaryLength)/float64(result.OriginalLength)))
	}

	fmt.Println(strings.Repeat("=", 80))
	return nil
}
