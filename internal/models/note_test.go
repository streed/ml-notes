package models

import (
	"testing"
	"time"
)

func TestNoteValidation(t *testing.T) {
	tests := []struct {
		name    string
		note    Note
		wantErr bool
	}{
		{
			name: "Valid note",
			note: Note{
				Title:   "Test Note",
				Content: "This is test content",
			},
			wantErr: false,
		},
		{
			name: "Empty title",
			note: Note{
				Title:   "",
				Content: "Content without title",
			},
			wantErr: false, // Title can be empty based on current implementation
		},
		{
			name: "Empty content",
			note: Note{
				Title:   "Title without content",
				Content: "",
			},
			wantErr: false, // Content can be empty based on current implementation
		},
		{
			name: "Note with ID",
			note: Note{
				ID:      1,
				Title:   "Test Note",
				Content: "Test content",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add validation logic if needed in the future
			// For now, just test that the struct can be created
			if tt.note.Title == "" && tt.note.Content == "" && tt.wantErr {
				t.Error("Expected error for empty note, but validation not implemented")
			}
		})
	}
}

func TestNoteTimestamps(t *testing.T) {
	now := time.Now()
	note := Note{
		ID:        1,
		Title:     "Test Note",
		Content:   "Test content",
		CreatedAt: now,
		UpdatedAt: now.Add(1 * time.Hour),
	}

	if note.CreatedAt.After(note.UpdatedAt) {
		t.Error("CreatedAt should not be after UpdatedAt")
	}

	if note.UpdatedAt.Sub(note.CreatedAt) != 1*time.Hour {
		t.Error("UpdatedAt should be 1 hour after CreatedAt")
	}
}

// TestNoteEmbedding removed as Note struct doesn't have Embedding field
// Embeddings are stored separately in the database

func TestNoteFields(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	note := Note{
		ID:        42,
		Title:     "Important Note",
		Content:   "This is very important content",
		CreatedAt: testTime,
		UpdatedAt: testTime.Add(24 * time.Hour),
	}

	if note.ID != 42 {
		t.Errorf("Expected ID 42, got %d", note.ID)
	}

	if note.Title != "Important Note" {
		t.Errorf("Expected title 'Important Note', got %s", note.Title)
	}

	if note.Content != "This is very important content" {
		t.Errorf("Expected specific content, got %s", note.Content)
	}

	if !note.CreatedAt.Equal(testTime) {
		t.Errorf("CreatedAt mismatch: expected %v, got %v", testTime, note.CreatedAt)
	}

	expectedUpdate := testTime.Add(24 * time.Hour)
	if !note.UpdatedAt.Equal(expectedUpdate) {
		t.Errorf("UpdatedAt mismatch: expected %v, got %v", expectedUpdate, note.UpdatedAt)
	}
}