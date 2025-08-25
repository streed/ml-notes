package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/autotag"
	"github.com/streed/ml-notes/internal/models"
)

var autoTagCmd = &cobra.Command{
	Use:   "auto-tag",
	Short: "Automatically generate tags for notes using AI",
	Long: `Use AI to analyze note content and automatically suggest or apply tags.
This feature uses your configured Ollama instance to intelligently analyze
note content and generate relevant organizational tags.

Examples:
  ml-notes auto-tag 123                    # Suggest tags for note 123
  ml-notes auto-tag 123 456 789           # Suggest tags for multiple notes
  ml-notes auto-tag --all                 # Suggest tags for all notes
  ml-notes auto-tag --apply 123           # Apply suggested tags automatically
  ml-notes auto-tag --all --apply         # Auto-tag all notes (use with caution!)
  ml-notes auto-tag --recent 10 --apply   # Auto-tag 10 most recent notes`,
	RunE: runAutoTag,
}

var (
	autoTagApply     bool
	autoTagAll       bool
	autoTagRecent    int
	autoTagOverwrite bool
	autoTagDryRun    bool
)

func init() {
	rootCmd.AddCommand(autoTagCmd)
	autoTagCmd.Flags().BoolVar(&autoTagApply, "apply", false, "Automatically apply suggested tags to notes")
	autoTagCmd.Flags().BoolVar(&autoTagAll, "all", false, "Process all notes in the database")
	autoTagCmd.Flags().IntVar(&autoTagRecent, "recent", 0, "Process N most recent notes")
	autoTagCmd.Flags().BoolVar(&autoTagOverwrite, "overwrite", false, "Overwrite existing tags (default: merge with existing)")
	autoTagCmd.Flags().BoolVar(&autoTagDryRun, "dry-run", false, "Show what would be tagged without applying changes")
}

func runAutoTag(_ *cobra.Command, args []string) error {
	// Create auto-tagger service
	autoTagger := autotag.NewAutoTagger(appConfig)

	// Check if auto-tagging is available
	if !autoTagger.IsAvailable() {
		return fmt.Errorf("auto-tagging is not available. Please ensure:\n" +
			"1. Summarization is enabled (ml-notes config set enable-summarization true)\n" +
			"2. Ollama is running and accessible\n" +
			"3. A summarization model is configured")
	}

	// Determine which notes to process
	var notes []*models.Note
	var err error

	if autoTagAll {
		fmt.Println("🔍 Loading all notes for auto-tagging...")
		notes, err = noteRepo.List(0, 0) // Get all notes
		if err != nil {
			return fmt.Errorf("failed to load notes: %w", err)
		}
	} else if autoTagRecent > 0 {
		fmt.Printf("🔍 Loading %d most recent notes for auto-tagging...\n", autoTagRecent)
		notes, err = noteRepo.List(autoTagRecent, 0)
		if err != nil {
			return fmt.Errorf("failed to load recent notes: %w", err)
		}
	} else if len(args) > 0 {
		// Process specific note IDs
		fmt.Printf("🔍 Loading %d specified notes for auto-tagging...\n", len(args))
		for _, arg := range args {
			id, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid note ID '%s': must be a number", arg)
			}

			note, err := noteRepo.GetByID(id)
			if err != nil {
				return fmt.Errorf("failed to get note %d: %w", id, err)
			}
			notes = append(notes, note)
		}
	} else {
		return fmt.Errorf("must specify note IDs, --all, or --recent N")
	}

	if len(notes) == 0 {
		fmt.Println("No notes found to process.")
		return nil
	}

	fmt.Printf("🤖 Processing %d notes with AI auto-tagging...\n", len(notes))

	if autoTagDryRun {
		fmt.Println("🧪 DRY RUN MODE - No changes will be applied")
	} else if autoTagApply {
		fmt.Println("⚡ APPLY MODE - Tags will be automatically applied")
	} else {
		fmt.Println("💡 SUGGESTION MODE - Tags will be suggested but not applied")
	}

	// Process notes for auto-tagging
	successCount := 0
	errorCount := 0

	for i, note := range notes {
		fmt.Printf("📝 Processing note %d/%d (ID: %d): %s\n", i+1, len(notes), note.ID, truncateTitle(note.Title, 50))

		// Get suggested tags
		suggestedTags, err := autoTagger.SuggestTags(note)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			errorCount++
			continue
		}

		if len(suggestedTags) == 0 {
			fmt.Printf("   ⚠️  No tags suggested\n")
			continue
		}

		// Show suggestions
		fmt.Printf("   🏷️  Suggested tags: %s\n", strings.Join(suggestedTags, ", "))

		// Show existing tags if any
		if len(note.Tags) > 0 {
			fmt.Printf("   🏷️  Existing tags: %s\n", strings.Join(note.Tags, ", "))
		}

		// Determine final tags
		var finalTags []string
		if autoTagOverwrite || len(note.Tags) == 0 {
			finalTags = suggestedTags
		} else {
			// Merge with existing tags, avoiding duplicates
			tagSet := make(map[string]bool)
			for _, tag := range note.Tags {
				tagSet[tag] = true
				finalTags = append(finalTags, tag)
			}
			for _, tag := range suggestedTags {
				if !tagSet[tag] {
					finalTags = append(finalTags, tag)
				}
			}
		}

		// Apply tags if requested and not in dry-run mode
		if autoTagApply && !autoTagDryRun {
			if err := noteRepo.UpdateTags(note.ID, finalTags); err != nil {
				fmt.Printf("   ❌ Failed to apply tags: %v\n", err)
				errorCount++
				continue
			}

			fmt.Printf("   ✅ Applied tags: %s\n", strings.Join(finalTags, ", "))
		} else if autoTagDryRun {
			fmt.Printf("   🧪 Would apply tags: %s\n", strings.Join(finalTags, ", "))
		}

		successCount++
		fmt.Println() // Empty line for readability
	}

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📊 Auto-tagging Summary:\n")
	fmt.Printf("   ✅ Successfully processed: %d notes\n", successCount)
	if errorCount > 0 {
		fmt.Printf("   ❌ Errors encountered: %d notes\n", errorCount)
	}

	if autoTagDryRun {
		fmt.Printf("\n🧪 This was a dry run. To apply the changes, run the same command without --dry-run\n")
	} else if !autoTagApply {
		fmt.Printf("\n💡 Tags were suggested but not applied. To apply automatically, use --apply flag\n")
	}

	return nil
}

// truncateTitle helper function to truncate long titles for display
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}
