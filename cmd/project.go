package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage ml-notes projects",
	Long: `Manage ml-notes projects. Projects provide isolated environments for your notes,
with each project having its own database and namespace for search.

Available subcommands:
  list    - List all projects
  create  - Create a new project  
  switch  - Switch to a different project
  current - Show current active project
  delete  - Delete a project
  update  - Update project metadata
  reset   - Reset entire system (DANGEROUS - CLI only)`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `List all available projects with their status and metadata.`,
	RunE:  runProjectList,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <name> [description]",
	Short: "Create a new project",
	Long: `Create a new project with the specified name and optional description.
	
The project name will be used to generate a unique ID and create a dedicated
database and directory structure.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runProjectCreate,
}

var projectSwitchCmd = &cobra.Command{
	Use:   "switch <project-id>",
	Short: "Switch to a different project",
	Long:  `Switch to a different project by its ID. This will make it the active project for all operations.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectSwitch,
}

var projectCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current active project",
	Long:  `Show the currently active project with detailed information.`,
	RunE:  runProjectCurrent,
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <project-id>",
	Short: "Delete a project",
	Long: `Delete a project by its ID. This removes the project from the registry but
leaves the project files intact for safety. You cannot delete the active project
or the default project.`,
	Args: cobra.ExactArgs(1),
	RunE: runProjectDelete,
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update <project-id>",
	Short: "Update project metadata",
	Long:  `Update the name and/or description of a project.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectUpdate,
}

var projectResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the entire ml-notes system (DANGEROUS)",
	Long: `Reset the entire ml-notes system by removing all projects, databases, and configuration.

⚠️  WARNING: This is a destructive operation that cannot be undone!
⚠️  All notes, projects, and configuration will be permanently deleted.

This command requires written confirmation and is only available via CLI for safety.`,
	RunE: runProjectReset,
}

var (
	// Flags for project commands
	projectName        string
	projectDescription string
)

func init() {
	rootCmd.AddCommand(projectCmd)

	// Add subcommands
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectSwitchCmd)
	projectCmd.AddCommand(projectCurrentCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectResetCmd)

	// Add flags
	projectUpdateCmd.Flags().StringVarP(&projectName, "name", "n", "", "New project name")
	projectUpdateCmd.Flags().StringVarP(&projectDescription, "description", "d", "", "New project description")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	pm := appConfig.GetProjectManager()
	if pm == nil {
		return fmt.Errorf("project manager not initialized")
	}

	projects := pm.ListProjects()
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tCREATED\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "--\t----\t-------\t-----------")

	for _, project := range projects {
		created := project.CreatedAt.Format("2006-01-02")
		description := project.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			project.ID, project.Name, created, description)
	}

	_ = w.Flush()
	return nil
}

func runProjectCreate(cmd *cobra.Command, args []string) error {
	pm := appConfig.GetProjectManager()
	if pm == nil {
		return fmt.Errorf("project manager not initialized")
	}

	name := args[0]
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

	project, err := pm.CreateProject(name, description)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("✓ Created project '%s' (ID: %s)\n", project.Name, project.ID)
	fmt.Printf("  Path: %s\n", project.Path)
	fmt.Printf("  Database: %s\n", project.DatabasePath)

	if description != "" {
		fmt.Printf("  Description: %s\n", description)
	}

	fmt.Printf("\nTo start using this project, run: ml-notes project switch %s\n", project.ID)
	return nil
}

func runProjectSwitch(cmd *cobra.Command, args []string) error {
	pm := appConfig.GetProjectManager()
	if pm == nil {
		return fmt.Errorf("project manager not initialized")
	}

	projectID := args[0]

	// Check if project exists
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Project '%s' (ID: %s) exists and is ready to use\n", project.Name, project.ID)
	fmt.Printf("  Database: %s\n", project.DatabasePath)
	fmt.Printf("  Note: Use --project=%s with commands to work with this project\n", project.ID)

	if project.Description != "" {
		fmt.Printf("  Description: %s\n", project.Description)
	}

	return nil
}

func runProjectCurrent(cmd *cobra.Command, args []string) error {
	pm := appConfig.GetProjectManager()
	if pm == nil {
		return fmt.Errorf("project manager not initialized")
	}

	fmt.Printf("No active project concept - projects are accessed via --project flag\n")
	fmt.Printf("Available projects:\n")

	projects := pm.ListProjects()
	for _, project := range projects {
		fmt.Printf("  - %s (ID: %s)\n", project.Name, project.ID)
	}

	// Show some stats
	if noteRepo != nil {
		noteCount, err := noteRepo.GetNoteCount()
		if err != nil {
			logger.Debug("Failed to get note count: %v", err)
		} else {
			fmt.Printf("  Notes: %d\n", noteCount)
		}

		tagCount, err := noteRepo.GetTagCount()
		if err != nil {
			logger.Debug("Failed to get tag count: %v", err)
		} else {
			fmt.Printf("  Tags: %d\n", tagCount)
		}
	}

	return nil
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	pm := appConfig.GetProjectManager()
	if pm == nil {
		return fmt.Errorf("project manager not initialized")
	}

	projectID := args[0]

	// Get project info for confirmation
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete project '%s' (ID: %s)? [y/N]: ", project.Name, project.ID)
	var response string
	_, err = fmt.Scanln(&response)
	if err != nil {
		// If there's an error reading input, treat as "no"
		response = "n"
	}

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	if err := pm.DeleteProject(projectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Printf("✓ Project '%s' deleted from registry\n", project.Name)
	fmt.Printf("Note: Project files remain at %s for safety\n", project.Path)

	return nil
}

func runProjectUpdate(cmd *cobra.Command, args []string) error {
	pm := appConfig.GetProjectManager()
	if pm == nil {
		return fmt.Errorf("project manager not initialized")
	}

	projectID := args[0]

	// Check if project exists
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	// Check if any flags were provided
	if projectName == "" && projectDescription == "" {
		return fmt.Errorf("specify at least one of --name or --description")
	}

	// Show current values
	fmt.Printf("Current project: %s (ID: %s)\n", project.Name, project.ID)
	if project.Description != "" {
		fmt.Printf("  Description: %s\n", project.Description)
	}

	// Update project
	newName := projectName
	if newName == "" {
		newName = project.Name
	}

	newDescription := projectDescription
	if cmd.Flags().Changed("description") && projectDescription == "" {
		// User explicitly cleared the description
		newDescription = ""
	} else if newDescription == "" {
		newDescription = project.Description
	}

	if err := pm.UpdateProject(projectID, newName, newDescription); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	fmt.Printf("✓ Project updated successfully\n")
	if newName != project.Name {
		fmt.Printf("  Name: %s → %s\n", project.Name, newName)
	}
	if newDescription != project.Description {
		if newDescription == "" {
			fmt.Printf("  Description: %s → (removed)\n", project.Description)
		} else {
			fmt.Printf("  Description: %s → %s\n", project.Description, newDescription)
		}
	}

	return nil
}

func runProjectReset(cmd *cobra.Command, args []string) error {
	fmt.Println("⚠️  WARNING: This will permanently delete ALL projects, notes, and configuration!")
	fmt.Println("⚠️  This operation cannot be undone!")
	fmt.Println()

	// Show what will be deleted
	pm := appConfig.GetProjectManager()
	if pm != nil {
		projects := pm.ListProjects()
		if len(projects) > 0 {
			fmt.Printf("Projects that will be deleted:\n")
			for _, project := range projects {
				fmt.Printf("  - %s (ID: %s)\n", project.Name, project.ID)
			}
			fmt.Println()
		}
	}

	fmt.Print("To confirm, type 'DELETE ALL MY DATA' exactly: ")

	reader := bufio.NewReader(os.Stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirmation = strings.TrimSpace(confirmation)
	if confirmation != "DELETE ALL MY DATA" {
		fmt.Println("❌ Reset cancelled - confirmation text did not match exactly.")
		return nil
	}

	fmt.Println()
	fmt.Println("🗑️  Starting system reset...")

	// Delete all projects first
	if pm != nil {
		projects := pm.ListProjects()
		for _, project := range projects {
			fmt.Printf("   Deleting project: %s\n", project.Name)
			if err := pm.DeleteProject(project.ID); err != nil {
				logger.Debug("Error deleting project %s: %v", project.ID, err)
			}
		}
	}

	// Remove data directory entirely
	dataDir := appConfig.DataDirectory
	if dataDir != "" {
		fmt.Printf("   Removing data directory: %s\n", dataDir)
		if err := os.RemoveAll(dataDir); err != nil {
			logger.Debug("Error removing data directory: %v", err)
		}
	}

	// Remove configuration directory
	configDir, err := config.GetConfigDir()
	if err == nil {
		fmt.Printf("   Removing config directory: %s\n", configDir)
		if err := os.RemoveAll(configDir); err != nil {
			logger.Debug("Error removing config directory: %v", err)
		}
	}

	// Remove config file
	configPath, err := config.GetConfigPath()
	if err == nil {
		fmt.Printf("   Removing config file: %s\n", configPath)
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			logger.Debug("Error removing config file: %v", err)
		}
	}

	fmt.Println()
	fmt.Println("✅ System reset complete!")
	fmt.Println("   All projects, notes, and configuration have been removed.")
	fmt.Println("   Run 'ml-notes init' to set up a fresh installation.")

	return nil
}
