package migrations

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/streed/ml-notes/internal/logger"
)

// Migration represents a single database migration
type Migration struct {
	ID          string                 // Unique identifier (e.g., "001_add_attachments")
	Description string                 // Human-readable description
	Up          func(tx *sql.Tx) error // Migration function
	Down        func(tx *sql.Tx) error // Rollback function (optional)
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *sql.DB) *MigrationRunner {
	return &MigrationRunner{
		db:         db,
		migrations: getAllMigrations(),
	}
}

// createMigrationsTable creates the migrations tracking table if it doesn't exist
func (mr *MigrationRunner) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id TEXT PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			checksum TEXT
		)
	`
	_, err := mr.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// getAppliedMigrations returns a set of migration IDs that have been applied
func (mr *MigrationRunner) getAppliedMigrations() (map[string]bool, error) {
	query := "SELECT id FROM schema_migrations"
	rows, err := mr.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan migration id: %w", err)
		}
		applied[id] = true
	}

	return applied, nil
}

// recordMigration records that a migration has been applied
func (mr *MigrationRunner) recordMigration(tx *sql.Tx, migration Migration) error {
	query := `
		INSERT INTO schema_migrations (id, description, applied_at) 
		VALUES (?, ?, ?)
	`
	_, err := tx.Exec(query, migration.ID, migration.Description, time.Now().UTC())
	return err
}

// RunMigrations runs all pending migrations
func (mr *MigrationRunner) RunMigrations() error {
	logger.Info("Starting database migrations...")

	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(); err != nil {
		return err
	}

	// Get already applied migrations
	applied, err := mr.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Sort migrations by ID to ensure they run in order
	sort.Slice(mr.migrations, func(i, j int) bool {
		return mr.migrations[i].ID < mr.migrations[j].ID
	})

	// Run pending migrations
	pendingCount := 0
	for _, migration := range mr.migrations {
		if applied[migration.ID] {
			logger.Debug("Migration %s already applied, skipping", migration.ID)
			continue
		}

		logger.Info("Running migration: %s - %s", migration.ID, migration.Description)

		// Start transaction
		tx, err := mr.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction for migration %s: %w", migration.ID, err)
		}

		// Run migration
		if err := migration.Up(tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Error("Failed to rollback transaction: %v", rollbackErr)
			}
			return fmt.Errorf("migration %s failed: %w", migration.ID, err)
		}

		// Record migration
		if err := mr.recordMigration(tx, migration); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Error("Failed to rollback transaction: %v", rollbackErr)
			}
			return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.ID, err)
		}

		pendingCount++
		logger.Info("Migration %s completed successfully", migration.ID)
	}

	if pendingCount == 0 {
		logger.Info("No pending migrations found - database is up to date")
	} else {
		logger.Info("Successfully applied %d migrations", pendingCount)
	}

	return nil
}

// GetMigrationStatus returns the status of all migrations
func (mr *MigrationRunner) GetMigrationStatus() ([]MigrationStatus, error) {
	applied, err := mr.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	var status []MigrationStatus
	for _, migration := range mr.migrations {
		status = append(status, MigrationStatus{
			ID:          migration.ID,
			Description: migration.Description,
			Applied:     applied[migration.ID],
		})
	}

	return status, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Applied     bool   `json:"applied"`
}

// RollbackMigration rolls back a specific migration (if supported)
func (mr *MigrationRunner) RollbackMigration(migrationID string) error {
	// Find the migration
	var targetMigration *Migration
	for _, migration := range mr.migrations {
		if migration.ID == migrationID {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %s not found", migrationID)
	}

	if targetMigration.Down == nil {
		return fmt.Errorf("migration %s does not support rollback", migrationID)
	}

	// Check if migration is applied
	applied, err := mr.getAppliedMigrations()
	if err != nil {
		return err
	}

	if !applied[migrationID] {
		return fmt.Errorf("migration %s is not applied", migrationID)
	}

	logger.Info("Rolling back migration: %s - %s", targetMigration.ID, targetMigration.Description)

	// Start transaction
	tx, err := mr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction for rollback %s: %w", migrationID, err)
	}

	// Run rollback
	if err := targetMigration.Down(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Error("Failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("rollback %s failed: %w", migrationID, err)
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE id = ?", migrationID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Error("Failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("failed to remove migration record %s: %w", migrationID, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", migrationID, err)
	}

	logger.Info("Migration %s rolled back successfully", migrationID)
	return nil
}
