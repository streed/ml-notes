package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/summarize"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [ids...]",
	Short: "Analyze notes with AI",
	Long: `Generate AI-powered analysis of one or more notes with reasoning process.
	
You can analyze:
- A single note by ID: ml-notes analyze 123
- Multiple notes by IDs: ml-notes analyze 1 2 3
- All notes: ml-notes analyze --all
- Recent notes: ml-notes analyze --recent 10

Custom analysis prompts:
- Focus on specific aspects: ml-notes analyze 123 -p "Focus on technical details"
- Extract insights: ml-notes analyze --all -p "What are the key themes and patterns?"
- Compare approaches: ml-notes analyze 1 2 3 -p "Compare and contrast these approaches"`,
	RunE: runAnalyze,
}

var (
	analyzeAll        bool
	analyzeRecent     int
	analyzeModel      string
	analyzePrompt     string
	writeBackCurrent  bool
	writeBackNew      bool
	writeBackTitle    string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVar(&analyzeAll, "all", false, "Analyze all notes")
	analyzeCmd.Flags().IntVar(&analyzeRecent, "recent", 0, "Analyze N most recent notes")
	analyzeCmd.Flags().StringVar(&analyzeModel, "model", "", "Override the analysis model")
	analyzeCmd.Flags().StringVarP(&analyzePrompt, "prompt", "p", "", "Custom analysis prompt (e.g., \"Focus on technical aspects\")")
	analyzeCmd.Flags().BoolVar(&writeBackCurrent, "write-back", false, "Write analysis back to the current note (only works with single note analysis)")
	analyzeCmd.Flags().BoolVar(&writeBackNew, "write-new", false, "Write analysis to a new note")
	analyzeCmd.Flags().StringVar(&writeBackTitle, "write-title", "", "Title for new note when using --write-new (defaults to 'Analysis of [original title]')")
}

func runAnalyze(_ *cobra.Command, args []string) error {
	if !appConfig.EnableSummarization {
		return fmt.Errorf("analysis is disabled in configuration")
	}

	var notes []*models.Note
	var err error

	// Determine which notes to analyze
	if analyzeAll {
		fmt.Println("Loading all notes...")
		notes, err = noteRepo.List(0, 0)
		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}
	} else if analyzeRecent > 0 {
		fmt.Printf("Loading %d most recent notes...\n", analyzeRecent)
		notes, err = noteRepo.List(analyzeRecent, 0)
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

	fmt.Printf("Analyzing %d note(s)...\n\n", len(notes))
	fmt.Println(strings.Repeat("=", 80))

	// Create analyzer
	analyzer := summarize.NewSummarizer(appConfig)
	if analyzeModel != "" {
		analyzer.SetModel(analyzeModel)
		fmt.Printf("Using model: %s\n\n", analyzeModel)
	} else if appConfig.SummarizationModel != "" {
		analyzer.SetModel(appConfig.SummarizationModel)
		fmt.Printf("Using model: %s\n\n", appConfig.SummarizationModel)
	}

	// Validate write-back flags
	if writeBackCurrent && writeBackNew {
		return fmt.Errorf("cannot use both --write-back and --write-new flags together")
	}
	if writeBackCurrent && len(notes) > 1 {
		return fmt.Errorf("--write-back can only be used with single note analysis")
	}

	// Generate analysis based on number of notes
	var analysisResult *summarize.SummaryResult
	if len(notes) == 1 {
		// Single note analysis
		note := notes[0]
		fmt.Printf("Note: %s (ID: %d)\n", note.Title, note.ID)
		fmt.Println(strings.Repeat("-", 80))

		result, err := analyzer.SummarizeNoteWithPrompt(note, analyzePrompt)
		if err != nil {
			return fmt.Errorf("failed to generate analysis: %w", err)
		}
		analysisResult = result

		fmt.Println("\n📝 Analysis:")
		fmt.Println(result.Summary)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("\n✨ Analysis generated using %s\n", result.Model)
		fmt.Printf("   Reduced from %d to %d characters (%.1f%% compression)\n",
			result.OriginalLength, result.SummaryLength,
			100.0*(1.0-float64(result.SummaryLength)/float64(result.OriginalLength)))

		// Handle write-back to current note
		if writeBackCurrent {
			fmt.Println("\n💾 Writing analysis back to current note...")
			analysisSection := fmt.Sprintf("\n\n---\n## Analysis\n\n%s\n\n*Analysis generated on %s using %s*",
				result.Summary,
				time.Now().Format("2006-01-02 15:04:05"),
				result.Model)
			
			newContent := note.Content + analysisSection
			_, err := noteRepo.UpdateByID(note.ID, note.Title, newContent)
			if err != nil {
				return fmt.Errorf("failed to update note with analysis: %w", err)
			}
			fmt.Printf("✅ Analysis written back to note %d\n", note.ID)
		}
	} else {
		// Multiple notes analysis
		fmt.Printf("Analyzing %d notes together...\n", len(notes))
		fmt.Println(strings.Repeat("-", 80))

		// Show titles of notes being analyzed
		fmt.Println("Notes included:")
		for i, note := range notes {
			fmt.Printf("  %d. [ID: %d] %s\n", i+1, note.ID, note.Title)
			if i >= 10 && len(notes) > 12 {
				fmt.Printf("  ... and %d more notes\n", len(notes)-11)
				break
			}
		}
		fmt.Println()

		result, err := analyzer.SummarizeNotesWithPrompt(notes, "", analyzePrompt)
		if err != nil {
			return fmt.Errorf("failed to generate analysis: %w", err)
		}
		analysisResult = result

		fmt.Println("📝 Combined Analysis:")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println(result.Summary)
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("\n✨ Analysis generated using %s\n", result.Model)
		fmt.Printf("   Reduced from %d to %d characters (%.1f%% compression)\n",
			result.OriginalLength, result.SummaryLength,
			100.0*(1.0-float64(result.SummaryLength)/float64(result.OriginalLength)))
	}

	// Handle write to new note
	if writeBackNew && analysisResult != nil {
		fmt.Println("\n💾 Creating new note with analysis...")
		
		var newTitle string
		if writeBackTitle != "" {
			newTitle = writeBackTitle
		} else {
			if len(notes) == 1 {
				newTitle = fmt.Sprintf("Analysis of %s", notes[0].Title)
			} else {
				newTitle = fmt.Sprintf("Analysis of %d Notes", len(notes))
			}
		}

		// Create content with metadata
		var sourceInfo string
		if len(notes) == 1 {
			sourceInfo = fmt.Sprintf("**Source:** Note %d - %s", notes[0].ID, notes[0].Title)
		} else {
			sourceInfo = "**Sources:**\n"
			for i, note := range notes {
				sourceInfo += fmt.Sprintf("- Note %d: %s\n", note.ID, note.Title)
				if i >= 10 && len(notes) > 12 {
					sourceInfo += fmt.Sprintf("- ... and %d more notes\n", len(notes)-11)
					break
				}
			}
		}

		newContent := fmt.Sprintf("%s\n\n---\n\n%s\n\n---\n\n*Analysis generated on %s using %s*",
			sourceInfo,
			analysisResult.Summary,
			time.Now().Format("2006-01-02 15:04:05"),
			analysisResult.Model)

		newNote, err := noteRepo.Create(newTitle, newContent)
		if err != nil {
			return fmt.Errorf("failed to create new note with analysis: %w", err)
		}

		// Index the note for vector search
		fullText := newTitle + " " + newContent
		if err := vectorSearch.IndexNote(newNote.ID, fullText); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index analysis note for vector search: %v\n", err)
		}

		fmt.Printf("✅ Analysis written to new note %d: %s\n", newNote.ID, newTitle)
	}

	fmt.Println(strings.Repeat("=", 80))
	return nil
}
