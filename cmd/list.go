package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes",
	Long:  `List all notes with their ID, title, and creation date.`,
	RunE:  runList,
}

var (
	listLimit  int
	listOffset int
	listShort  bool
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 20, "Maximum number of notes to display")
	listCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "Number of notes to skip")
	listCmd.Flags().BoolVarP(&listShort, "short", "s", false, "Show only ID and title")
}

func runList(cmd *cobra.Command, args []string) error {
	notes, err := noteRepo.List(listLimit, listOffset)
	if err != nil {
		return fmt.Errorf("failed to list notes: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	fmt.Printf("Found %d notes:\n\n", len(notes))

	for _, note := range notes {
		if listShort {
			fmt.Printf("[%d] %s\n", note.ID, note.Title)
		} else {
			fmt.Printf("ID: %d\n", note.ID)
			fmt.Printf("Title: %s\n", note.Title)
			fmt.Printf("Created: %s\n", formatTime(note.CreatedAt))

			// Show preview of content
			preview := note.Content
			if len(preview) > 100 {
				preview = preview[:97] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			fmt.Printf("Preview: %s\n", preview)
			fmt.Println(strings.Repeat("-", 60))
		}
	}

	return nil
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02 15:04")
	}
}
