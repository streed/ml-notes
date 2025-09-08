package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
)

type DB struct {
	conn *sql.DB
	cfg  *config.Config
}

func New(cfg *config.Config) (*DB, error) {
	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.GetDatabasePath())
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	logger.Debug("Database path: %s", cfg.GetDatabasePath())

	conn, err := sql.Open("sqlite3", cfg.GetDatabasePath())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn, cfg: cfg}
	if err := db.initialize(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

func (db *DB) initialize() error {
	// Create notes table
	_, err := db.conn.Exec(`
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

	// Create tags table
	_, err = db.conn.Exec(`
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
	_, err = db.conn.Exec(`
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

	// Create indexes for better query performance
	_, err = db.conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_note_tags_note_id ON note_tags(note_id);
		CREATE INDEX IF NOT EXISTS idx_note_tags_tag_id ON note_tags(tag_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}

// MigrateFromPerProjectDatabases migrates data from per-project databases to the single multi-tenant database
func (db *DB) MigrateFromPerProjectDatabases(projectManager interface{}) error {
	// Type assertion to get project manager methods
	type ProjectLister interface {
		ListProjectsForMigration() []interface{}
	}
	
	pm, ok := projectManager.(ProjectLister)
	if !ok {
		logger.Debug("Project manager does not implement ProjectLister interface, skipping migration")
		return nil
	}
	
	projects := pm.ListProjectsForMigration()
	
	for _, proj := range projects {
		// Type assertion to get project fields
		type ProjectData interface {
			GetID() string
			GetName() string
			GetDescription() string
			GetDatabasePath() string
		}
		
		project, ok := proj.(ProjectData)
		if !ok {
			continue
		}
		
		if err := db.migrateProjectDatabase(project); err != nil {
			return fmt.Errorf("failed to migrate project %s: %w", project.GetID(), err)
		}
	}
	
	return nil
}

// migrateProjectDatabase migrates a single project database
func (db *DB) migrateProjectDatabase(project interface{}) error {
	type ProjectData interface {
		GetID() string
		GetName() string
		GetDescription() string
		GetDatabasePath() string
	}
	
	p, ok := project.(ProjectData)
	if !ok {
		return fmt.Errorf("invalid project data")
	}
	
	// Check if project database exists
	if _, err := os.Stat(p.GetDatabasePath()); os.IsNotExist(err) {
		logger.Debug("Project database does not exist: %s", p.GetDatabasePath())
		return nil // Nothing to migrate
	}
	
	// Open project database
	projectDB, err := sql.Open("sqlite3", p.GetDatabasePath())
	if err != nil {
		return fmt.Errorf("failed to open project database: %w", err)
	}
	defer projectDB.Close()
	
	// First ensure the project exists in the projects table
	_, err = db.conn.Exec(`
		INSERT OR IGNORE INTO projects (id, name, description)
		VALUES (?, ?, ?)
	`, p.GetID(), p.GetName(), p.GetDescription())
	if err != nil {
		return fmt.Errorf("failed to create project record: %w", err)
	}
	
	// Migrate notes
	if err := db.migrateNotesFromProject(projectDB, p.GetID()); err != nil {
		return fmt.Errorf("failed to migrate notes: %w", err)
	}
	
	// Migrate tags and note_tags
	if err := db.migrateTagsFromProject(projectDB, p.GetID()); err != nil {
		return fmt.Errorf("failed to migrate tags: %w", err)
	}
	
	logger.Debug("Successfully migrated project %s from %s", p.GetID(), p.GetDatabasePath())
	return nil
}

// migrateNotesFromProject migrates notes from a project database
func (db *DB) migrateNotesFromProject(projectDB *sql.DB, projectID string) error {
	// Query notes from project database
	rows, err := projectDB.Query(`
		SELECT id, title, content, created_at, updated_at 
		FROM notes ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("failed to query project notes: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var id int
		var title, content, createdAt, updatedAt string
		
		if err := rows.Scan(&id, &title, &content, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("failed to scan note: %w", err)
		}
		
		// Insert into main database
		_, err = db.conn.Exec(`
			INSERT OR IGNORE INTO notes (title, content, created_at, updated_at)
			VALUES (?, ?, ?, ?)
		`, title, content, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert migrated note: %w", err)
		}
	}
	
	return rows.Err()
}

// migrateTagsFromProject migrates tags and note_tags from a project database
func (db *DB) migrateTagsFromProject(projectDB *sql.DB, projectID string) error {
	// First migrate tags
	tagRows, err := projectDB.Query(`
		SELECT id, name, created_at FROM tags ORDER BY id
	`)
	if err != nil {
		// Tags table might not exist in old databases
		logger.Debug("No tags table found in project database, skipping tag migration")
		return nil
	}
	defer tagRows.Close()
	
	// Map old tag IDs to new tag IDs
	tagIDMap := make(map[int]int)
	
	for tagRows.Next() {
		var oldTagID int
		var name, createdAt string
		
		if err := tagRows.Scan(&oldTagID, &name, &createdAt); err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		
		// Insert or get existing tag
		var newTagID int
		err = db.conn.QueryRow(`
			SELECT id FROM tags WHERE name = ?
		`, name).Scan(&newTagID)
		
		if err == sql.ErrNoRows {
			// Tag doesn't exist, create it
			result, err := db.conn.Exec(`
				INSERT INTO tags (name, created_at) VALUES (?, ?)
			`, name, createdAt)
			if err != nil {
				return fmt.Errorf("failed to insert tag: %w", err)
			}
			
			id, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get tag insert id: %w", err)
			}
			newTagID = int(id)
		} else if err != nil {
			return fmt.Errorf("failed to query existing tag: %w", err)
		}
		
		tagIDMap[oldTagID] = newTagID
	}
	
	if err = tagRows.Err(); err != nil {
		return fmt.Errorf("error iterating tag rows: %w", err)
	}
	
	// Now migrate note_tags relationships
	noteTagRows, err := projectDB.Query(`
		SELECT nt.note_id, nt.tag_id, nt.created_at, n.title
		FROM note_tags nt
		JOIN notes n ON nt.note_id = n.id
		ORDER BY nt.note_id, nt.tag_id
	`)
	if err != nil {
		// note_tags table might not exist
		logger.Debug("No note_tags table found in project database, skipping note_tags migration")
		return nil
	}
	defer noteTagRows.Close()
	
	for noteTagRows.Next() {
		var oldNoteID, oldTagID int
		var createdAt, noteTitle string
		
		if err := noteTagRows.Scan(&oldNoteID, &oldTagID, &createdAt, &noteTitle); err != nil {
			return fmt.Errorf("failed to scan note_tag: %w", err)
		}
		
		// Find the new note ID by title
		var newNoteID int
		err = db.conn.QueryRow(`
			SELECT id FROM notes WHERE title = ?
		`, noteTitle).Scan(&newNoteID)
		if err != nil {
			logger.Debug("Could not find migrated note for note_tag relationship: %s", noteTitle)
			continue
		}
		
		// Get new tag ID from mapping
		newTagID, exists := tagIDMap[oldTagID]
		if !exists {
			logger.Debug("Could not find migrated tag for note_tag relationship: %d", oldTagID)
			continue
		}
		
		// Insert note_tag relationship
		_, err = db.conn.Exec(`
			INSERT OR IGNORE INTO note_tags (note_id, tag_id, created_at)
			VALUES (?, ?, ?)
		`, newNoteID, newTagID, createdAt)
		if err != nil {
			return fmt.Errorf("failed to insert note_tag relationship: %w", err)
		}
	}
	
	return noteTagRows.Err()
}
