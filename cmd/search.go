package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/summarize"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes",
	Long:  `Search notes using text matching or vector similarity search.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearch,
}

var (
	searchLimit     int
	useVector       bool
	searchShort     bool
	searchSummarize bool
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results")
	searchCmd.Flags().BoolVarP(&useVector, "vector", "v", false, "Use vector similarity search")
	searchCmd.Flags().BoolVarP(&searchShort, "short", "s", false, "Show only ID and title")
	searchCmd.Flags().BoolVar(&searchSummarize, "summarize", false, "Generate a summary of search results")
}

func runSearch(_ *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	var notes []*models.Note
	var err error

	if useVector {
		fmt.Printf("Performing vector similarity search for: %s\n\n", query)
		notes, err = vectorSearch.SearchSimilar(query, searchLimit)
		if err != nil {
			return fmt.Errorf("vector search failed: %w", err)
		}
	} else {
		fmt.Printf("Performing text search for: %s\n\n", query)
		notes, err = noteRepo.Search(query)
		if err != nil {
			return fmt.Errorf("text search failed: %w", err)
		}
		// Apply limit to text search results
		if len(notes) > searchLimit {
			notes = notes[:searchLimit]
		}
	}

	if len(notes) == 0 {
		fmt.Println("No matching notes found.")
		return nil
	}

	fmt.Printf("Found %d matching notes:\n\n", len(notes))

	// Generate summary if requested
	if searchSummarize && appConfig.EnableSummarization {
		fmt.Println("Generating summary of search results...")
		fmt.Println(strings.Repeat("=", 80))

		summarizer := summarize.NewSummarizer(appConfig)
		if appConfig.SummarizationModel != "" {
			summarizer.SetModel(appConfig.SummarizationModel)
		}

		result, err := summarizer.SummarizeNotes(notes, query)
		if err != nil {
			logger.Error("Failed to generate summary: %v", err)
			fmt.Printf("Warning: Could not generate summary: %v\n", err)
		} else {
			fmt.Println("\n📝 Summary of Search Results:")
			fmt.Println(strings.Repeat("-", 80))
			fmt.Println(result.Summary)
			fmt.Println(strings.Repeat("-", 80))
			fmt.Printf("\n✨ Summary generated using %s\n", result.Model)
			fmt.Printf("   Reduced from %d to %d characters (%.1f%% compression)\n\n",
				result.OriginalLength, result.SummaryLength,
				100.0*(1.0-float64(result.SummaryLength)/float64(result.OriginalLength)))
		}
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("\nDetailed Results:")
	}

	for i, note := range notes {
		if searchShort {
			fmt.Printf("[%d] %s\n", note.ID, note.Title)
		} else {
			if useVector {
				fmt.Printf("Match %d:\n", i+1)
			}
			fmt.Printf("ID: %d\n", note.ID)
			fmt.Printf("Title: %s\n", note.Title)
			fmt.Printf("Created: %s\n", formatTime(note.CreatedAt))

			// Show preview with query context
			preview := note.Content
			if len(preview) > 150 {
				// Try to show context around the query if doing text search
				if !useVector {
					lowerContent := strings.ToLower(note.Content)
					lowerQuery := strings.ToLower(query)
					idx := strings.Index(lowerContent, lowerQuery)
					if idx >= 0 {
						start := idx - 50
						if start < 0 {
							start = 0
						}
						end := idx + len(query) + 100
						if end > len(note.Content) {
							end = len(note.Content)
						}
						preview = "..." + note.Content[start:end] + "..."
					} else {
						preview = preview[:147] + "..."
					}
				} else {
					preview = preview[:147] + "..."
				}
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			fmt.Printf("Preview: %s\n", preview)
			fmt.Println(strings.Repeat("-", 60))
		}
	}

	return nil
}
