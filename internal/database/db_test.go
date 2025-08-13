package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/streed/ml-notes/internal/config"
)

func setupTestDB(t *testing.T) (*DB, string) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	cfg := &config.Config{
		DatabasePath:       dbPath,
		DataDirectory:      tempDir,
		VectorDimensions:   3,
		EnableVectorSearch: true,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db, dbPath
}

func TestNew(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer db.Close()

	// Check that database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Check that we can query the database
	var version string
	err := db.conn.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		t.Errorf("Failed to query SQLite version: %v", err)
	}

	if version == "" {
		t.Error("SQLite version should not be empty")
	}
}

func TestDatabaseInitialization(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	// Check that notes table was created
	var tableExists int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='notes'",
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf("Failed to check for notes table: %v", err)
	}
	if tableExists != 1 {
		t.Error("Notes table should exist")
	}

	// Check that note_embeddings table was created
	err = db.conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='note_embeddings'",
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf("Failed to check for note_embeddings table: %v", err)
	}
	if tableExists != 1 {
		t.Error("Note_embeddings table should exist")
	}
}

func TestClose(t *testing.T) {
	db, _ := setupTestDB(t)

	err := db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Try to query after close - should fail
	var version string
	err = db.conn.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err == nil {
		t.Error("Expected error when querying closed database")
	}
}

func TestConn(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	conn := db.Conn()
	if conn == nil {
		t.Error("Conn() should return non-nil connection")
	}

	// Verify the connection works
	var test int
	err := conn.QueryRow("SELECT 1").Scan(&test)
	if err != nil {
		t.Errorf("Failed to use connection: %v", err)
	}

	if test != 1 {
		t.Errorf("Expected 1, got %d", test)
	}
}

func TestDatabaseWithEmptyConfig(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		DataDirectory: tempDir,
		// DatabasePath will be generated
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database with empty config: %v", err)
	}
	defer db.Close()

	expectedPath := filepath.Join(tempDir, "notes.db")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("Database file should be created at default location")
	}
}

func TestDatabaseCreatesDirectories(t *testing.T) {
	tempDir := t.TempDir()
	deepPath := filepath.Join(tempDir, "level1", "level2", "level3")
	dbPath := filepath.Join(deepPath, "test.db")

	cfg := &config.Config{
		DatabasePath:     dbPath,
		DataDirectory:    tempDir,
		VectorDimensions: 384,
	}

	db, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create database in nested directory: %v", err)
	}
	defer db.Close()

	// Check that all directories were created
	if _, err := os.Stat(deepPath); os.IsNotExist(err) {
		t.Error("Nested directories should be created")
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file should be created")
	}
}
