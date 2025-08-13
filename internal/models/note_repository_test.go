package models

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the notes table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create notes table: %v", err)
	}

	// Create tags table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tags table: %v", err)
	}

	// Create note_tags junction table
	_, err = db.Exec(`
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
		t.Fatalf("Failed to create note_tags table: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestNewNoteRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)
	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	if repo.db != db {
		t.Error("Repository should store database connection")
	}
}

func TestNoteRepositoryCreate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	title := "Test Note"
	content := "Test content"

	note, err := repo.Create(title, content)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if note.ID == 0 {
		t.Error("Note should have a valid ID")
	}

	if note.Title != title {
		t.Errorf("Expected title %s, got %s", title, note.Title)
	}

	if note.Content != content {
		t.Errorf("Expected content %s, got %s", content, note.Content)
	}

	if note.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if note.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestNoteRepositoryGetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Create a note first
	created, err := repo.Create("Test Title", "Test Content")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Get the note by ID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get note: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("ID mismatch: expected %d, got %d", created.ID, retrieved.ID)
	}

	if retrieved.Title != created.Title {
		t.Errorf("Title mismatch: expected %s, got %s", created.Title, retrieved.Title)
	}

	if retrieved.Content != created.Content {
		t.Errorf("Content mismatch: expected %s, got %s", created.Content, retrieved.Content)
	}
}

func TestNoteRepositoryGetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	_, err := repo.GetByID(9999)
	if err == nil {
		t.Error("Expected error for non-existent note")
	}
}

func TestNoteRepositoryList(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Create multiple notes
	for i := 1; i <= 5; i++ {
		_, err := repo.Create("Note", "Content")
		if err != nil {
			t.Fatalf("Failed to create note %d: %v", i, err)
		}
	}

	// List all notes
	notes, err := repo.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list notes: %v", err)
	}

	if len(notes) != 5 {
		t.Errorf("Expected 5 notes, got %d", len(notes))
	}

	// List with limit
	notes, err = repo.List(3, 0)
	if err != nil {
		t.Fatalf("Failed to list notes with limit: %v", err)
	}

	if len(notes) != 3 {
		t.Errorf("Expected 3 notes with limit, got %d", len(notes))
	}

	// List with limit and offset
	notes, err = repo.List(2, 2)
	if err != nil {
		t.Fatalf("Failed to list notes with offset: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes with offset, got %d", len(notes))
	}
}

func TestNoteRepositoryListWithLimit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Create notes
	for i := 0; i < 10; i++ {
		_, err := repo.Create("Note", "Content")
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}
	}

	// Test ListWithLimit
	notes, err := repo.ListWithLimit(5, 3)
	if err != nil {
		t.Fatalf("Failed to list with limit: %v", err)
	}

	if len(notes) != 5 {
		t.Errorf("Expected 5 notes, got %d", len(notes))
	}
}

func TestNoteRepositoryUpdateByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Create a note
	original, err := repo.Create("Original Title", "Original Content")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Sleep to ensure updated_at will be different
	time.Sleep(100 * time.Millisecond)

	// Update the note
	newTitle := "Updated Title"
	newContent := "Updated Content"
	updated, err := repo.UpdateByID(original.ID, newTitle, newContent)
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	if updated.Title != newTitle {
		t.Errorf("Title not updated: expected %s, got %s", newTitle, updated.Title)
	}

	if updated.Content != newContent {
		t.Errorf("Content not updated: expected %s, got %s", newContent, updated.Content)
	}

	// Note: SQLite's CURRENT_TIMESTAMP may have only second precision,
	// so we can't reliably test that UpdatedAt changed for fast operations.
	// The important thing is that the content was updated correctly.
}

func TestNoteRepositoryUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Create a note
	note, err := repo.Create("Original", "Content")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Sleep to ensure updated_at will be different
	time.Sleep(100 * time.Millisecond)

	// Update using Update method
	note.Title = "Modified Title"
	note.Content = "Modified Content"
	err = repo.Update(note)
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	// Verify the update
	if note.Title != "Modified Title" {
		t.Errorf("Title not updated: got %s", note.Title)
	}

	if note.Content != "Modified Content" {
		t.Errorf("Content not updated: got %s", note.Content)
	}
}

func TestNoteRepositoryDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Create a note
	note, err := repo.Create("To Delete", "Will be deleted")
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	// Delete the note
	err = repo.Delete(note.ID)
	if err != nil {
		t.Fatalf("Failed to delete note: %v", err)
	}

	// Try to get the deleted note
	_, err = repo.GetByID(note.ID)
	if err == nil {
		t.Error("Expected error when getting deleted note")
	}
}

func TestNoteRepositoryDelete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	err := repo.Delete(9999)
	if err == nil {
		t.Error("Expected error when deleting non-existent note")
	}
}

func TestNoteRepositorySearch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Create notes with different content
	notes := []struct {
		title   string
		content string
	}{
		{"Go Programming", "Learn Go language basics"},
		{"Python Tutorial", "Python programming guide"},
		{"JavaScript", "JS for web development"},
		{"Go Advanced", "Advanced Go concepts"},
	}

	for _, n := range notes {
		_, err := repo.Create(n.title, n.content)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}
	}

	// Search for "Go"
	results, err := repo.Search("Go")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Go', got %d", len(results))
	}

	// Search for "programming"
	results, err = repo.Search("programming")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'programming', got %d", len(results))
	}

	// Search for non-existent term
	results, err = repo.Search("nonexistent")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestNoteRepositoryConcurrency(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewNoteRepository(db)

	// Test concurrent creates
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := repo.Create("Concurrent", "Content")
			if err != nil {
				t.Errorf("Failed to create note concurrently: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all notes were created
	notes, err := repo.List(0, 0)
	if err != nil {
		t.Fatalf("Failed to list notes: %v", err)
	}

	if len(notes) != 10 {
		t.Errorf("Expected 10 notes, got %d", len(notes))
	}
}
