package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/streed/ml-notes/internal/migrations"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management",
	Long: `Manage database migrations and schema changes.

This command provides utilities to check migration status and manage database schema changes.`,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of database migrations",
	Long: `Display which database migrations have been applied and which are pending.

This helps you understand the current state of your database schema.`,
	RunE: showMigrationStatus,
}

var migrateRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run pending database migrations",
	Long: `Manually run any pending database migrations.

Note: Migrations are automatically run when the application starts, so this command
is typically only needed for troubleshooting or advanced use cases.`,
	RunE: runMigrations,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateRunCmd)
}

func showMigrationStatus(cmd *cobra.Command, args []string) error {
	// Use existing database connection
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create migration runner
	migrationRunner := migrations.NewMigrationRunner(db.Conn())

	// Get migration status
	status, err := migrationRunner.GetMigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Display results in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "MIGRATION ID\tSTATUS\tDESCRIPTION\n")
	fmt.Fprintf(w, "------------\t------\t-----------\n")

	for _, migration := range status {
		statusText := "PENDING"
		if migration.Applied {
			statusText = "APPLIED"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", migration.ID, statusText, migration.Description)
	}

	w.Flush()

	// Summary
	appliedCount := 0
	for _, migration := range status {
		if migration.Applied {
			appliedCount++
		}
	}

	fmt.Printf("\nTotal migrations: %d\n", len(status))
	fmt.Printf("Applied: %d\n", appliedCount)
	fmt.Printf("Pending: %d\n", len(status)-appliedCount)

	return nil
}

func runMigrations(cmd *cobra.Command, args []string) error {
	// Use existing database connection
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create migration runner
	migrationRunner := migrations.NewMigrationRunner(db.Conn())

	// Run migrations
	if err := migrationRunner.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Migration run completed successfully!")
	return nil
}
