package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/logger"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [note IDs...]",
	Short: "Delete one or more notes",
	Long: `Delete notes by their IDs. 
	
You can delete a single note or multiple notes at once.
By default, you will be prompted for confirmation before deletion.
Use --force to skip the confirmation prompt.
Use --all to delete all notes (no IDs required).`,
	Args:    validateDeleteArgs,
	Aliases: []string{"rm", "remove"},
	RunE:    runDelete,
}

var (
	forceDelete bool
	deleteAll   bool
)

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Delete all notes (use with caution!)")
}

// validateDeleteArgs ensures proper arguments are provided
func validateDeleteArgs(cmd *cobra.Command, args []string) error {
	// Get the flag value directly from the command
	allFlag, _ := cmd.Flags().GetBool("all")
	
	// If --all flag is set, no IDs should be provided
	if allFlag {
		if len(args) > 0 {
			return fmt.Errorf("cannot specify note IDs when using --all flag")
		}
		return nil
	}
	
	// If --all is not set, at least one ID is required
	if len(args) < 1 {
		return fmt.Errorf("requires at least one note ID (or use --all to delete all notes)")
	}
	
	return nil
}

func runDelete(_ *cobra.Command, args []string) error {
	// Handle delete all case
	if deleteAll {
		return deleteAllNotes()
	}

	// Parse note IDs
	noteIDs := make([]int, 0, len(args))
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("invalid note ID '%s': must be a number", arg)
		}
		noteIDs = append(noteIDs, id)
	}

	// Get notes to display what will be deleted
	notesToDelete := make(map[int]string)
	for _, id := range noteIDs {
		note, err := noteRepo.GetByID(id)
		if err != nil {
			logger.Error("Note with ID %d not found: %v", id, err)
			fmt.Printf("Warning: Note with ID %d not found\n", id)
			continue
		}
		notesToDelete[id] = note.Title
	}

	if len(notesToDelete) == 0 {
		fmt.Println("No valid notes to delete.")
		return nil
	}

	// Show what will be deleted
	fmt.Println("The following notes will be deleted:")
	fmt.Println(strings.Repeat("-", 60))
	for id, title := range notesToDelete {
		fmt.Printf("  [%d] %s\n", id, title)
	}
	fmt.Println(strings.Repeat("-", 60))

	// Confirm deletion unless force flag is set
	if !forceDelete {
		if !confirmDeletion(len(notesToDelete)) {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete notes
	successCount := 0
	failCount := 0
	for id := range notesToDelete {
		if err := noteRepo.Delete(id); err != nil {
			logger.Error("Failed to delete note %d: %v", id, err)
			fmt.Printf("✗ Failed to delete note %d: %v\n", id, err)
			failCount++
		} else {
			fmt.Printf("✓ Deleted note %d: %s\n", id, notesToDelete[id])
			successCount++
			
			// Also remove from vector index if enabled
			if appConfig.EnableVectorSearch && vectorSearch != nil {
				// Vector search cleanup happens automatically via CASCADE in database
				logger.Debug("Note %d removed from vector index", id)
			}
		}
	}

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	if failCount == 0 {
		fmt.Printf("Successfully deleted %d note(s).\n", successCount)
	} else {
		fmt.Printf("Deleted %d note(s), failed to delete %d note(s).\n", successCount, failCount)
	}

	return nil
}

func deleteAllNotes() error {
	// Get all notes
	allNotes, err := noteRepo.List(100000, 0) // Get all notes (up to 100k)
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	noteCount := len(allNotes)
	if noteCount == 0 {
		fmt.Println("No notes to delete.")
		return nil
	}

	// Show warning
	fmt.Printf("⚠️  WARNING: This will delete ALL %d notes!\n", noteCount)
	fmt.Println("This action cannot be undone.")
	fmt.Println(strings.Repeat("=", 60))

	// Require explicit confirmation for delete all
	if !forceDelete {
		fmt.Printf("Type 'DELETE ALL %d NOTES' to confirm: ", noteCount)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		
		expectedResponse := fmt.Sprintf("DELETE ALL %d NOTES", noteCount)
		if response != expectedResponse {
			fmt.Println("Confirmation text did not match. Deletion cancelled.")
			return nil
		}
	} else {
		// Even with force, require confirmation for delete all
		if !confirmDeletion(noteCount) {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete all notes
	successCount := 0
	failCount := 0
	for _, note := range allNotes {
		if err := noteRepo.Delete(note.ID); err != nil {
			logger.Error("Failed to delete note %d: %v", note.ID, err)
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	if failCount == 0 {
		fmt.Printf("Successfully deleted all %d notes.\n", successCount)
	} else {
		fmt.Printf("Deleted %d notes, failed to delete %d notes.\n", successCount, failCount)
	}

	return nil
}

func confirmDeletion(count int) bool {
	var prompt string
	if count == 1 {
		prompt = "Are you sure you want to delete this note? (y/N): "
	} else {
		prompt = fmt.Sprintf("Are you sure you want to delete %d notes? (y/N): ", count)
	}

	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}