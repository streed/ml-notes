package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new note",
	Long:  `Add a new note with a title and content. You can provide content via stdin or interactively.`,
	RunE:  runAdd,
}

var (
	title   string
	content string
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&title, "title", "t", "", "Note title (required)")
	addCmd.Flags().StringVarP(&content, "content", "c", "", "Note content")
	addCmd.MarkFlagRequired("title")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if content == "" {
		// Read content interactively or from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped to stdin
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			content = strings.Join(lines, "\n")
		} else {
			// Interactive mode
			fmt.Println("Enter note content (press Ctrl+D when finished):")
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			content = strings.Join(lines, "\n")
		}
	}

	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	note, err := noteRepo.Create(title, content)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	// Index the note for vector search
	fullText := title + " " + content
	if err := vectorSearch.IndexNote(note.ID, fullText); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to index note for vector search: %v\n", err)
	}

	fmt.Printf("Note created successfully!\n")
	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}