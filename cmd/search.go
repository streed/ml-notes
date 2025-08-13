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
	Long:  `Search notes using text matching, vector similarity search, or tag search.
	
When using vector search, only the most similar note is returned by default.
Use --limit to get more results.

You can search by tags only using --tags flag without providing a query.`,
	Args:  cobra.ArbitraryArgs,
	RunE:  runSearch,
}

var (
	searchLimit       int
	useVector         bool
	searchShort       bool
	searchAnalyze     bool
	searchShowDetails bool
	searchPrompt      string
	searchTags        []string
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", -1, "Maximum number of results (-1 for default: 1 for vector, 10 for text)")
	searchCmd.Flags().BoolVarP(&useVector, "vector", "v", false, "Use vector similarity search (returns top match by default)")
	searchCmd.Flags().BoolVarP(&searchShort, "short", "s", false, "Show only ID and title")
	searchCmd.Flags().BoolVar(&searchAnalyze, "analyze", false, "Generate an analysis of search results (hides details unless --show-details is used)")
	searchCmd.Flags().BoolVar(&searchShowDetails, "show-details", false, "Show detailed results even when analyzing")
	searchCmd.Flags().StringVarP(&searchPrompt, "prompt", "p", "", "Custom analysis prompt when using --analyze (e.g., \"Focus on technical aspects\")")
	searchCmd.Flags().StringSliceVarP(&searchTags, "tags", "T", []string{}, "Search for notes with any of these tags (comma-separated)")
}

func runSearch(_ *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	// Check if we're doing a tag-only search
	if len(searchTags) > 0 && query == "" {
		// Tag-only search
		return runTagSearch()
	}
	
	// Validate that we have either a query or tags
	if query == "" && len(searchTags) == 0 {
		return fmt.Errorf("must provide either a search query or tags (use --tags flag)")
	}

	var notes []*models.Note
	var err error

	// Set default limit based on search type if not specified
	effectiveLimit := searchLimit
	if searchLimit == -1 {
		if useVector {
			effectiveLimit = 1 // Default to top result for vector search
		} else {
			effectiveLimit = 10 // Default to 10 results for text search
		}
	}

	if useVector {
		fmt.Printf("Performing vector similarity search for: %s\n", query)
		if effectiveLimit == 1 {
			fmt.Println("(Returning the most similar note)")
		} else {
			fmt.Printf("(Returning top %d most similar notes)\n", effectiveLimit)
		}
		fmt.Println()
		notes, err = vectorSearch.SearchSimilar(query, effectiveLimit)
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
		if len(notes) > effectiveLimit {
			notes = notes[:effectiveLimit]
		}
	}

	if len(notes) == 0 {
		fmt.Println("No matching notes found.")
		return nil
	}

	if len(notes) == 1 {
		if useVector {
			fmt.Println("Found the most similar note:")
		} else {
			fmt.Println("Found 1 matching note:")
		}
	} else {
		fmt.Printf("Found %d matching notes:\n", len(notes))
	}
	fmt.Println()

	// Generate analysis if requested
	if searchAnalyze && appConfig.EnableSummarization {
		fmt.Println("Generating analysis of search results...")
		fmt.Println(strings.Repeat("=", 80))

		analyzer := summarize.NewSummarizer(appConfig)
		if appConfig.SummarizationModel != "" {
			analyzer.SetModel(appConfig.SummarizationModel)
		}

		result, err := analyzer.SummarizeNotesWithPrompt(notes, query, searchPrompt)
		if err != nil {
			logger.Error("Failed to generate analysis: %v", err)
			fmt.Printf("Warning: Could not generate analysis: %v\n", err)
			// If analysis fails, fall back to showing detailed results
		} else {
			fmt.Println("\n📝 Analysis of Search Results:")
			fmt.Println(strings.Repeat("-", 80))
			fmt.Println(result.Summary)
			fmt.Println(strings.Repeat("-", 80))
			fmt.Printf("\n✨ Analysis generated using %s\n", result.Model)
			fmt.Printf("   Reduced from %d to %d characters (%.1f%% compression)\n",
				result.OriginalLength, result.SummaryLength,
				100.0*(1.0-float64(result.SummaryLength)/float64(result.OriginalLength)))
			fmt.Println(strings.Repeat("=", 80))
			
			// When analysis is successful, only show details if explicitly requested
			if !searchShowDetails {
				return nil
			}
			fmt.Println("\nDetailed Results:")
		}
	}

	// Show detailed results if not analyzing, if analysis failed, or if explicitly requested
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
			if len(note.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
			}
			fmt.Println(strings.Repeat("-", 60))
		}
	}

	return nil
}

// runTagSearch performs a search based only on tags
func runTagSearch() error {
	fmt.Printf("Searching for notes with tags: %s\n\n", strings.Join(searchTags, ", "))
	
	notes, err := noteRepo.SearchByTags(searchTags)
	if err != nil {
		return fmt.Errorf("tag search failed: %w", err)
	}
	
	// Apply limit if specified
	effectiveLimit := searchLimit
	if searchLimit == -1 {
		effectiveLimit = 10 // Default limit for tag search
	}
	if len(notes) > effectiveLimit {
		notes = notes[:effectiveLimit]
	}
	
	if len(notes) == 0 {
		fmt.Println("No notes found with the specified tags.")
		return nil
	}
	
	if len(notes) == 1 {
		fmt.Println("Found 1 note:")
	} else {
		fmt.Printf("Found %d notes:\n", len(notes))
	}
	fmt.Println()
	
	// Display results
	for _, note := range notes {
		if searchShort {
			fmt.Printf("[%d] %s\n", note.ID, note.Title)
		} else {
			fmt.Printf("ID: %d\n", note.ID)
			fmt.Printf("Title: %s\n", note.Title)
			fmt.Printf("Created: %s\n", formatTime(note.CreatedAt))
			if len(note.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
			}
			
			// Show preview
			preview := note.Content
			if len(preview) > 150 {
				preview = preview[:147] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			fmt.Printf("Preview: %s\n", preview)
			fmt.Println(strings.Repeat("-", 60))
		}
	}
	
	return nil
}
