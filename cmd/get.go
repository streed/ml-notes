package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a note by ID",
	Long:  `Display the full content of a note by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	note, err := noteRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	fmt.Printf("================================================================================\n")
	fmt.Printf("ID: %d\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	fmt.Printf("Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", note.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("================================================================================\n\n")
	fmt.Println(note.Content)
	fmt.Println()

	return nil
}