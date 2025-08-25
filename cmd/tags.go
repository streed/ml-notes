package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage note tags",
	Long: `Manage tags for notes. Tags help organize and categorize your notes.

Commands:
  list          List all tags
  add <id>      Add tags to a note
  remove <id>   Remove tags from a note
  set <id>      Set/replace all tags for a note`,
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	Long:  `List all tags in the system with their usage count.`,
	RunE:  runTagsList,
}

var tagsAddCmd = &cobra.Command{
	Use:   "add <note-id>",
	Short: "Add tags to a note",
	Long: `Add one or more tags to an existing note.
	
Example:
  ml-notes tags add 123 -T research,ai,machine-learning`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsAdd,
}

var tagsRemoveCmd = &cobra.Command{
	Use:   "remove <note-id>",
	Short: "Remove tags from a note",
	Long: `Remove specific tags from a note.
	
Example:
  ml-notes tags remove 123 -T outdated,old`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsRemove,
}

var tagsSetCmd = &cobra.Command{
	Use:   "set <note-id>",
	Short: "Set/replace all tags for a note",
	Long: `Replace all existing tags for a note with the specified tags.
	
Example:
  ml-notes tags set 123 -T research,updated,final`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsSet,
}

var (
	tagsList []string
)

func init() {
	rootCmd.AddCommand(tagsCmd)

	// Add subcommands
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsAddCmd)
	tagsCmd.AddCommand(tagsRemoveCmd)
	tagsCmd.AddCommand(tagsSetCmd)

	// Add flags for tag operations
	tagsAddCmd.Flags().StringSliceVarP(&tagsList, "tags", "T", []string{}, "Tags to add (comma-separated)")
	tagsRemoveCmd.Flags().StringSliceVarP(&tagsList, "tags", "T", []string{}, "Tags to remove (comma-separated)")
	tagsSetCmd.Flags().StringSliceVarP(&tagsList, "tags", "T", []string{}, "Tags to set (comma-separated)")

	// Mark tags flag as required for operations
	_ = tagsAddCmd.MarkFlagRequired("tags")
	_ = tagsRemoveCmd.MarkFlagRequired("tags")
	_ = tagsSetCmd.MarkFlagRequired("tags")
}

func runTagsList(_ *cobra.Command, _ []string) error {
	tags, err := noteRepo.GetAllTags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	fmt.Printf("Found %d tags:\n\n", len(tags))
	for _, tag := range tags {
		fmt.Printf("• %s\n", tag.Name)
	}

	return nil
}

func runTagsAdd(_ *cobra.Command, args []string) error {
	noteID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	// Get current note to verify it exists
	note, err := noteRepo.GetByID(noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	// Combine existing tags with new ones (avoiding duplicates)
	existingTags := make(map[string]bool)
	for _, tag := range note.Tags {
		existingTags[tag] = true
	}

	var newTags []string
	for _, tag := range tagsList {
		if !existingTags[tag] {
			newTags = append(newTags, tag)
		}
	}

	if len(newTags) == 0 {
		fmt.Println("All specified tags are already present on this note.")
		return nil
	}

	// Add new tags to existing ones
	allTags := append(note.Tags, newTags...)

	err = noteRepo.UpdateTags(noteID, allTags)
	if err != nil {
		return fmt.Errorf("failed to add tags: %w", err)
	}

	fmt.Printf("Added tags to note %d: %s\n", noteID, strings.Join(newTags, ", "))
	fmt.Printf("Note now has tags: %s\n", strings.Join(allTags, ", "))

	return nil
}

func runTagsRemove(_ *cobra.Command, args []string) error {
	noteID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	// Get current note to verify it exists
	note, err := noteRepo.GetByID(noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	// Remove specified tags
	tagsToRemove := make(map[string]bool)
	for _, tag := range tagsList {
		tagsToRemove[tag] = true
	}

	var remainingTags []string
	var removedTags []string
	for _, tag := range note.Tags {
		if tagsToRemove[tag] {
			removedTags = append(removedTags, tag)
		} else {
			remainingTags = append(remainingTags, tag)
		}
	}

	if len(removedTags) == 0 {
		fmt.Println("None of the specified tags were found on this note.")
		return nil
	}

	err = noteRepo.UpdateTags(noteID, remainingTags)
	if err != nil {
		return fmt.Errorf("failed to remove tags: %w", err)
	}

	fmt.Printf("Removed tags from note %d: %s\n", noteID, strings.Join(removedTags, ", "))
	if len(remainingTags) > 0 {
		fmt.Printf("Remaining tags: %s\n", strings.Join(remainingTags, ", "))
	} else {
		fmt.Println("Note now has no tags.")
	}

	return nil
}

func runTagsSet(_ *cobra.Command, args []string) error {
	noteID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid note ID: %s", args[0])
	}

	// Verify note exists
	_, err = noteRepo.GetByID(noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	err = noteRepo.UpdateTags(noteID, tagsList)
	if err != nil {
		return fmt.Errorf("failed to set tags: %w", err)
	}

	if len(tagsList) > 0 {
		fmt.Printf("Set tags for note %d: %s\n", noteID, strings.Join(tagsList, ", "))
	} else {
		fmt.Printf("Removed all tags from note %d\n", noteID)
	}

	return nil
}
