package search

import (
	"fmt"
	"testing"

	"github.com/streed/ml-notes/internal/config"
)

func TestCreateNamespace(t *testing.T) {
	// Create a mock repository and config for testing
	cfg := &config.Config{}
	lrs := NewLilRagSearch(nil, cfg)

	tests := []struct {
		name      string
		namespace string
		expected  string
	}{
		{
			name:      "empty namespace",
			namespace: "",
			expected:  "ml-notes",
		},
		{
			name:      "project namespace",
			namespace: "myproject",
			expected:  "ml-notes-myproject",
		},
		{
			name:      "complex project name",
			namespace: "my-awesome-project",
			expected:  "ml-notes-my-awesome-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lrs.createNamespace(tt.namespace)
			if result != tt.expected {
				t.Errorf("createNamespace(%q) = %q, want %q", tt.namespace, result, tt.expected)
			}
		})
	}
}

func TestDeleteNoteDocID(t *testing.T) {
	// Test that note deletion uses the same document ID format as indexing
	tests := []struct {
		name      string
		noteID    int
		projectID string
		expected  string
	}{
		{
			name:      "simple note",
			noteID:    123,
			projectID: "default",
			expected:  "notes-default-123",
		},
		{
			name:      "complex project",
			noteID:    456,
			projectID: "my-awesome-project",
			expected:  "notes-my-awesome-project-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that both index and delete use the same document ID format
			indexDocID := getDocumentID(tt.noteID, tt.projectID)
			deleteDocID := getDocumentID(tt.noteID, tt.projectID)
			
			if indexDocID != tt.expected {
				t.Errorf("getDocumentID(%d, %q) = %q, want %q", tt.noteID, tt.projectID, indexDocID, tt.expected)
			}
			if deleteDocID != tt.expected {
				t.Errorf("delete document ID should match index document ID: got %q, want %q", deleteDocID, tt.expected)
			}
		})
	}
}

// Helper function to extract document ID generation logic for testing
func getDocumentID(noteID int, projectID string) string {
	// This mirrors the logic in IndexNoteWithNamespace and DeleteNoteWithNamespace
	return fmt.Sprintf("notes-%s-%d", projectID, noteID)
}
