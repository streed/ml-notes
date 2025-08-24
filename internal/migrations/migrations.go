package migrations

import (
	"database/sql"
	"fmt"
)

// getAllMigrations returns all available migrations in order
func getAllMigrations() []Migration {
	return []Migration{
		{
			ID:          "000_initial_schema",
			Description: "Create initial database schema with notes, tags, and embeddings",
			Up:          migration000Up,
			Down:        migration000Down,
		},
		{
			ID:          "001_add_file_attachments",
			Description: "Add support for file attachments to notes",
			Up:          migration001Up,
			Down:        migration001Down,
		},
		// Add new migrations here in chronological order
	}
}

// migration001Up adds file attachment support
func migration001Up(tx *sql.Tx) error {
	// Check if note_attachments table already exists
	var tableExists bool
	err := tx.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='note_attachments'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if note_attachments table exists: %w", err)
	}

	if tableExists {
		// Table already exists, check if it has all required columns
		return ensureNoteAttachmentsSchema(tx)
	}

	// Create the note_attachments table
	_, err = tx.Exec(`
		CREATE TABLE note_attachments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			note_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			original_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			file_path TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create note_attachments table: %w", err)
	}

	// Create index on note_attachments for better query performance
	_, err = tx.Exec(`
		CREATE INDEX idx_note_attachments_note_id ON note_attachments(note_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create note_attachments index: %w", err)
	}

	return nil
}

// migration001Down removes file attachment support
func migration001Down(tx *sql.Tx) error {
	// Drop the index first
	_, err := tx.Exec("DROP INDEX IF EXISTS idx_note_attachments_note_id")
	if err != nil {
		return fmt.Errorf("failed to drop note_attachments index: %w", err)
	}

	// Drop the table
	_, err = tx.Exec("DROP TABLE IF EXISTS note_attachments")
	if err != nil {
		return fmt.Errorf("failed to drop note_attachments table: %w", err)
	}

	return nil
}

// ensureNoteAttachmentsSchema ensures the note_attachments table has all required columns
func ensureNoteAttachmentsSchema(tx *sql.Tx) error {
	// Get existing columns
	rows, err := tx.Query("PRAGMA table_info(note_attachments)")
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	existingColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk bool
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return fmt.Errorf("failed to scan column info: %w", err)
		}
		existingColumns[name] = true
	}

	// Required columns for note_attachments table
	requiredColumns := map[string]string{
		"id":            "INTEGER PRIMARY KEY AUTOINCREMENT",
		"note_id":       "INTEGER NOT NULL",
		"filename":      "TEXT NOT NULL",
		"original_name": "TEXT NOT NULL",
		"mime_type":     "TEXT NOT NULL",
		"file_size":     "INTEGER NOT NULL",
		"file_path":     "TEXT NOT NULL",
		"created_at":    "DATETIME DEFAULT CURRENT_TIMESTAMP",
	}

	// Add missing columns
	for columnName, columnDef := range requiredColumns {
		if !existingColumns[columnName] {
			// SQLite doesn't support adding PRIMARY KEY or FOREIGN KEY columns
			// If we're missing critical columns, we need to recreate the table
			if columnName == "id" {
				return recreateNoteAttachmentsTable(tx, existingColumns)
			}

			// Add the missing column
			alterSQL := fmt.Sprintf("ALTER TABLE note_attachments ADD COLUMN %s %s", columnName, columnDef)
			_, err = tx.Exec(alterSQL)
			if err != nil {
				return fmt.Errorf("failed to add column %s: %w", columnName, err)
			}
		}
	}

	// Ensure index exists
	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_note_attachments_note_id ON note_attachments(note_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create note_attachments index: %w", err)
	}

	return nil
}

// recreateNoteAttachmentsTable recreates the table with proper schema when major changes are needed
func recreateNoteAttachmentsTable(tx *sql.Tx, existingColumns map[string]bool) error {
	// First, create a temporary table with the new schema
	_, err := tx.Exec(`
		CREATE TABLE note_attachments_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			note_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			original_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			file_path TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create temporary table: %w", err)
	}

	// Copy data from old table to new table, handling missing columns
	copyColumns := []string{}
	insertPlaceholders := []string{}

	columnMappings := map[string]string{
		"id":            "id",
		"note_id":       "note_id",
		"filename":      "filename",
		"original_name": "original_name",
		"mime_type":     "mime_type",
		"file_size":     "file_size",
		"file_path":     "file_path",
		"created_at":    "created_at",
	}

	for newCol, oldCol := range columnMappings {
		if existingColumns[oldCol] {
			copyColumns = append(copyColumns, oldCol)
			insertPlaceholders = append(insertPlaceholders, newCol)
		}
	}

	if len(copyColumns) > 0 {
		copySQL := fmt.Sprintf(
			"INSERT INTO note_attachments_new (%s) SELECT %s FROM note_attachments",
			joinStrings(insertPlaceholders, ", "),
			joinStrings(copyColumns, ", "),
		)
		_, err = tx.Exec(copySQL)
		if err != nil {
			return fmt.Errorf("failed to copy data to new table: %w", err)
		}
	}

	// Drop the old table
	_, err = tx.Exec("DROP TABLE note_attachments")
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Rename the new table
	_, err = tx.Exec("ALTER TABLE note_attachments_new RENAME TO note_attachments")
	if err != nil {
		return fmt.Errorf("failed to rename new table: %w", err)
	}

	// Create index
	_, err = tx.Exec(`
		CREATE INDEX idx_note_attachments_note_id ON note_attachments(note_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index on recreated table: %w", err)
	}

	return nil
}

// migration000Up creates the initial database schema
func migration000Up(tx *sql.Tx) error {
	// Create notes table
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %w", err)
	}

	// Create embeddings table for vector search
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS note_embeddings (
			id INTEGER PRIMARY KEY,
			note_id INTEGER NOT NULL,
			embedding BLOB,
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create embeddings table: %w", err)
	}

	// Create tags table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	// Create note_tags junction table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS note_tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			note_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
			UNIQUE(note_id, tag_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create note_tags table: %w", err)
	}

	// Create indexes on note_tags for better query performance
	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_note_tags_note_id ON note_tags(note_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create note_tags note_id index: %w", err)
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_note_tags_tag_id ON note_tags(tag_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create note_tags tag_id index: %w", err)
	}

	return nil
}

// migration000Down drops the initial schema
func migration000Down(tx *sql.Tx) error {
	// Drop tables in reverse order of creation (to handle foreign keys)
	tables := []string{"note_tags", "tags", "note_embeddings", "notes"}

	for _, table := range tables {
		_, err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
