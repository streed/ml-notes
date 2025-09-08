package search

import "github.com/streed/ml-notes/internal/models"

// SearchProvider interface defines methods for indexing and searching notes
type SearchProvider interface {
	IndexNote(noteID int, content string) error
	SearchSimilar(query string, limit int) ([]*models.Note, error)
}
