package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/logger"
)

type DB struct {
	conn *sql.DB
	cfg  *config.Config
}

func New(cfg *config.Config) (*DB, error) {
	// Initialize sqlite-vec extension
	sqlite_vec.Auto()
	logger.Debug("Initialized sqlite-vec extension")

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
	// The sqlite-vec extension is now automatically loaded via the Go bindings
	// Test if vec0 is available
	var vecVersion string
	err := db.conn.QueryRow("SELECT vec_version()").Scan(&vecVersion)
	if err == nil {
		logger.Info("sqlite-vec version %s loaded", vecVersion)
	} else {
		logger.Debug("sqlite-vec not available: %v", err)
	}

	// Create notes table
	_, err = db.conn.Exec(`
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
	_, err = db.conn.Exec(`
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

	// Create virtual table for vector similarity search using vec0
	// We'll use the configured dimensions or default to 384
	dimensions := db.cfg.VectorDimensions
	if dimensions == 0 {
		dimensions = 384
	}

	_, err = db.conn.Exec(fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS vec_notes USING vec0(
			note_id INTEGER PRIMARY KEY,
			embedding float[%d]
		)
	`, dimensions))
	if err != nil {
		// Log but don't fail - we can still use fallback search
		logger.Warn("Vector table creation failed (vec0 may not be available): %v", err)
	} else {
		logger.Debug("Created vec_notes table with %d dimensions", dimensions)
	}

	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}
