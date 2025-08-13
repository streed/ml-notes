package cmd

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
)

var editCmd = &cobra.Command{
	Use:   "edit <note-id>",
	Short: "Edit an existing note",
	Long: `Edit a note in your default editor.
	
The note will open in your $EDITOR (or vi if not set).
After editing, if the content has changed, the note will be saved and reindexed.

The editor will show the note in this format:
  Title: [note title]
  ---
  [note content]

You can edit both the title and content. The first line after "Title: " becomes the new title.
Everything after the "---" separator becomes the content.`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var (
	editTitle   bool
	editContent bool
	editor      string
)

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().BoolVarP(&editTitle, "title", "t", false, "Edit title only")
	editCmd.Flags().BoolVarP(&editContent, "content", "c", false, "Edit content only")
	editCmd.Flags().StringVarP(&editor, "editor", "e", "", "Specify editor to use (overrides $EDITOR)")
}

func runEdit(_ *cobra.Command, args []string) error {
	// Parse note ID
	noteID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID '%s': must be a number", args[0])
	}

	// Get the note
	note, err := noteRepo.GetByID(noteID)
	if err != nil {
		return fmt.Errorf("failed to get note %d: %w", noteID, err)
	}

	// Store original content for comparison
	originalTitle := note.Title
	originalContent := note.Content
	originalHash := hashContent(note.Title, note.Content)

	// Determine what to edit
	var editedTitle, editedContent string
	
	if editTitle && !editContent {
		// Edit title only
		editedTitle, err = editInEditor(note.Title, noteID, true)
		if err != nil {
			return fmt.Errorf("failed to edit title: %w", err)
		}
		editedContent = note.Content
	} else if editContent && !editTitle {
		// Edit content only
		editedContent, err = editInEditor(note.Content, noteID, false)
		if err != nil {
			return fmt.Errorf("failed to edit content: %w", err)
		}
		editedTitle = note.Title
	} else {
		// Edit both (default)
		editedTitle, editedContent, err = editFullNote(note)
		if err != nil {
			return fmt.Errorf("failed to edit note: %w", err)
		}
	}

	// Check if content changed
	newHash := hashContent(editedTitle, editedContent)
	if originalHash == newHash {
		fmt.Println("No changes detected.")
		return nil
	}

	// Update the note
	note.Title = editedTitle
	note.Content = editedContent

	if err := noteRepo.Update(note); err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	// Reindex if content changed and vector search is enabled
	if appConfig.EnableVectorSearch && vectorSearch != nil {
		fmt.Println("Reindexing note for vector search...")
		fullText := editedTitle + " " + editedContent
		if err := vectorSearch.IndexNote(noteID, fullText); err != nil {
			logger.Error("Failed to reindex note %d: %v", noteID, err)
			fmt.Printf("Warning: Failed to reindex note for vector search: %v\n", err)
		} else {
			fmt.Println("✓ Note reindexed successfully")
		}
	}

	// Show what changed
	fmt.Println("\n✓ Note updated successfully")
	if originalTitle != editedTitle {
		fmt.Printf("  Title changed from: %s\n", originalTitle)
		fmt.Printf("                  to: %s\n", editedTitle)
	}
	if originalContent != editedContent {
		contentChangeSize := len(editedContent) - len(originalContent)
		if contentChangeSize > 0 {
			fmt.Printf("  Content increased by %d characters\n", contentChangeSize)
		} else if contentChangeSize < 0 {
			fmt.Printf("  Content decreased by %d characters\n", -contentChangeSize)
		} else {
			fmt.Println("  Content modified (same length)")
		}
	}

	return nil
}

// editFullNote opens the full note (title and content) in an editor
func editFullNote(note *models.Note) (string, string, error) {
	// Format note for editing
	content := fmt.Sprintf("Title: %s\n---\n%s", note.Title, note.Content)
	
	// Create temp file
	tempFile, err := os.CreateTemp("", fmt.Sprintf("ml-notes-%d-*.md", note.ID))
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write content to temp file
	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		return "", "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Open in editor
	if err := openInEditor(tempFile.Name()); err != nil {
		return "", "", err
	}

	// Read edited content
	editedBytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", "", fmt.Errorf("failed to read edited file: %w", err)
	}

	// Parse the edited content
	editedContent := string(editedBytes)
	lines := strings.Split(editedContent, "\n")
	
	// Find title and content separator
	var title string
	var contentStartIdx int
	
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, "Title: ") {
			title = strings.TrimPrefix(line, "Title: ")
		} else if line == "---" {
			contentStartIdx = i + 1
			break
		}
	}

	// Extract content (everything after the separator)
	var contentLines []string
	if contentStartIdx > 0 && contentStartIdx < len(lines) {
		contentLines = lines[contentStartIdx:]
	}
	content = strings.Join(contentLines, "\n")
	content = strings.TrimSpace(content)

	// If no valid format found, treat entire content as note content
	if title == "" && contentStartIdx == 0 {
		// User might have deleted the format, treat all as content
		title = note.Title // Keep original title
		content = strings.TrimSpace(editedContent)
	}

	return title, content, nil
}

// editInEditor opens text in an editor and returns the edited content
func editInEditor(text string, noteID int, isTitle bool) (string, error) {
	suffix := "content"
	if isTitle {
		suffix = "title"
	}
	
	// Create temp file
	tempFile, err := os.CreateTemp("", fmt.Sprintf("ml-notes-%d-%s-*.txt", noteID, suffix))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write content to temp file
	if _, err := tempFile.WriteString(text); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Open in editor
	if err := openInEditor(tempFile.Name()); err != nil {
		return "", err
	}

	// Read edited content
	editedBytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	return strings.TrimSpace(string(editedBytes)), nil
}

// openInEditor opens a file in the user's editor
func openInEditor(filename string) error {
	// Determine which editor to use - in order of preference:
	// 1. --editor flag
	// 2. Config editor setting
	// 3. EDITOR environment variable
	// 4. VISUAL environment variable
	// 5. Auto-detection of common editors
	editorCmd := editor
	if editorCmd == "" && appConfig != nil {
		editorCmd = appConfig.Editor
	}
	if editorCmd == "" {
		editorCmd = os.Getenv("EDITOR")
	}
	if editorCmd == "" {
		editorCmd = os.Getenv("VISUAL")
	}
	if editorCmd == "" {
		// Try to find a common editor
		for _, e := range []string{"vim", "vi", "nano", "emacs", "code", "subl", "atom"} {
			if _, err := exec.LookPath(e); err == nil {
				editorCmd = e
				break
			}
		}
	}
	if editorCmd == "" {
		return fmt.Errorf("no editor found. Set $EDITOR environment variable, use --editor flag, or configure with: ml-notes config set editor <editor>")
	}

	logger.Debug("Opening file in editor: %s %s", editorCmd, filename)

	// Handle editors that might have arguments (e.g., "code --wait")
	parts := strings.Fields(editorCmd)
	cmd := exec.Command(parts[0], append(parts[1:], filename)...)
	
	// Connect to terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run editor
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run editor %s: %w", editorCmd, err)
	}

	return nil
}

// hashContent creates a hash of the title and content for change detection
func hashContent(title, content string) string {
	h := md5.New()
	h.Write([]byte(title))
	h.Write([]byte("\n---\n"))
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// updateNote handles the database update and reindexing
func updateNote(note *models.Note) error {
	// Update in database
	if err := noteRepo.Update(note); err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	// Reindex for vector search if enabled
	if appConfig.EnableVectorSearch && vectorSearch != nil {
		fullText := note.Title + " " + note.Content
		if err := vectorSearch.IndexNote(note.ID, fullText); err != nil {
			logger.Error("Failed to reindex note %d: %v", note.ID, err)
			// Don't fail the update, just warn
			fmt.Printf("Warning: Failed to reindex note for vector search: %v\n", err)
		}
	}

	return nil
}