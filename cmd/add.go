package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/autotag"
	interrors "github.com/streed/ml-notes/internal/errors"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
	"github.com/streed/ml-notes/internal/search"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new note",
	Long: `Add a new note with a title and content.

Content can be provided in several ways:
1. Via --content flag: ml-notes add -t "Title" -c "Content"
2. Via stdin: echo "Content" | ml-notes add -t "Title"
3. Via editor (default): ml-notes add -t "Title"
4. Via editor explicitly: ml-notes add -t "Title" -e

When no content is provided and a terminal is available, your $EDITOR will open.
Set $EDITOR environment variable or use --editor-cmd to specify your preferred editor.`,
	RunE: runAdd,
}

var (
	title      string
	content    string
	useEditor  bool
	editorName string
	tags       []string
	autoTag    bool
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&title, "title", "t", "", "Note title (required)")
	addCmd.Flags().StringVarP(&content, "content", "c", "", "Note content")
	addCmd.Flags().StringSliceVarP(&tags, "tags", "T", []string{}, "Tags for the note (comma-separated)")
	addCmd.Flags().BoolVar(&autoTag, "auto-tag", false, "Automatically generate tags using AI")
	addCmd.Flags().BoolVarP(&useEditor, "editor", "e", false, "Use editor for content input")
	addCmd.Flags().StringVar(&editorName, "editor-cmd", "", "Specify editor to use (overrides $EDITOR)")
	_ = addCmd.MarkFlagRequired("title")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if content == "" {
		// Check if we should use editor
		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if useEditor && !isPiped {
			// Use editor for content input
			var err error
			content, err = getContentFromEditor(title)
			if err != nil {
				return fmt.Errorf("failed to get content from editor: %w", err)
			}
		} else if isPiped {
			// Data is being piped to stdin
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			content = strings.Join(lines, "\n")
		} else {
			// Interactive mode - check if terminal is available
			if isTerminalAvailable() {
				// Use editor by default if terminal is available
				var err error
				content, err = getContentFromEditor(title)
				if err != nil {
					// Fall back to stdin input if editor fails
					logger.Debug("Editor failed, falling back to stdin input: %v", err)
					fmt.Println("Enter note content (press Ctrl+D when finished):")
					scanner := bufio.NewScanner(os.Stdin)
					var lines []string
					for scanner.Scan() {
						lines = append(lines, scanner.Text())
					}
					content = strings.Join(lines, "\n")
				}
			} else {
				// No terminal, use stdin
				fmt.Println("Enter note content (press Ctrl+D when finished):")
				scanner := bufio.NewScanner(os.Stdin)
				var lines []string
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
				}
				content = strings.Join(lines, "\n")
			}
		}
	}

	if content == "" {
		return interrors.ErrEmptyContent
	}

	// Create note with tags if provided
	var note *models.Note
	var err error
	if len(tags) > 0 {
		note, err = noteRepo.CreateWithTags(title, content, tags)
	} else {
		note, err = noteRepo.Create(title, content)
	}
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	// Auto-tag if requested
	if autoTag {
		fmt.Println("🤖 Generating AI tags...")
		autoTagger := autotag.NewAutoTagger(appConfig)

		if autoTagger.IsAvailable() {
			suggestedTags, err := autoTagger.SuggestTags(note)
			if err != nil {
				fmt.Printf("⚠️  Auto-tagging failed: %v\n", err)
			} else if len(suggestedTags) > 0 {
				// Merge with existing tags
				allTags := note.Tags
				tagSet := make(map[string]bool)
				for _, tag := range allTags {
					tagSet[tag] = true
				}
				for _, tag := range suggestedTags {
					if !tagSet[tag] {
						allTags = append(allTags, tag)
					}
				}

				// Update note with auto-generated tags
				if err := noteRepo.UpdateTags(note.ID, allTags); err != nil {
					fmt.Printf("⚠️  Failed to apply auto-tags: %v\n", err)
				} else {
					note.Tags = allTags // Update for display
					fmt.Printf("🏷️  Auto-generated tags: %s\n", strings.Join(suggestedTags, ", "))
				}
			} else {
				fmt.Println("🏷️  No auto-tags generated")
			}
		} else {
			fmt.Printf("⚠️  Auto-tagging unavailable. Please ensure summarization is enabled and Ollama is running.\n")
		}
	}

	// Index the note for semantic search
	fullText := title + " " + content

	// Use namespace-aware indexing if available
	if lilragSearch, ok := vectorSearch.(*search.LilRagSearch); ok {
		namespace := getCurrentProjectNamespace()
		if err := lilragSearch.IndexNoteWithNamespace(note.ID, fullText, namespace, "default"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index note for semantic search: %v\n", err)
		}
	} else {
		if err := vectorSearch.IndexNote(note.ID, fullText); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index note for semantic search: %v\n", err)
		}
	}

	fmt.Printf("Note created successfully!\n")
	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	if len(note.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
	}
	fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

// getContentFromEditor opens an editor for the user to input content
func getContentFromEditor(noteTitle string) (string, error) {
	// Create temp file with helpful template
	tempFile, err := os.CreateTemp("", "ml-notes-new-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write template to temp file
	template := fmt.Sprintf(`# %s

[Write your note content here]

<!-- 
  Save and close the editor when done.
  To cancel, exit without saving.
-->`, noteTitle)

	if _, err := tempFile.WriteString(template); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Open in editor (reuse logic from edit command)
	if err := openEditor(tempFile.Name()); err != nil {
		return "", err
	}

	// Read edited content
	editedBytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	editedContent := string(editedBytes)

	// Remove the template comments if unchanged
	if strings.Contains(editedContent, "[Write your note content here]") {
		return "", fmt.Errorf("no content provided (template unchanged)")
	}

	// Clean up the content (remove template lines if present)
	lines := strings.Split(editedContent, "\n")
	var contentLines []string
	for _, line := range lines {
		// Skip title line and comment lines
		if strings.HasPrefix(line, "# "+noteTitle) {
			continue
		}
		if strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "-->") {
			continue
		}
		if strings.TrimSpace(line) == "[Write your note content here]" {
			continue
		}
		contentLines = append(contentLines, line)
	}

	finalContent := strings.TrimSpace(strings.Join(contentLines, "\n"))
	return finalContent, nil
}

// openEditor opens a file in the user's preferred editor
func openEditor(filename string) error {
	// Determine which editor to use - in order of preference:
	// 1. --editor-cmd flag
	// 2. Config editor setting
	// 3. EDITOR environment variable
	// 4. VISUAL environment variable
	// 5. Auto-detection of common editors
	editorCmd := editorName
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
		return fmt.Errorf("no editor found. Set $EDITOR environment variable, use --editor-cmd flag, or configure with: ml-notes config set editor <editor>")
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

// isTerminalAvailable checks if we're running in an interactive terminal
func isTerminalAvailable() bool {
	// Check if stdin is a terminal
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if it's a character device (terminal)
	return (stat.Mode() & os.ModeCharDevice) != 0
}
