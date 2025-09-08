package search

import (
	"fmt"
	"strconv"

	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/lilrag"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
)

type LilRagSearch struct {
	client *lilrag.Client
	repo   *models.NoteRepository
	cfg    *config.Config
}

func NewLilRagSearch(repo *models.NoteRepository, cfg *config.Config) *LilRagSearch {
	return &LilRagSearch{
		client: lilrag.NewClient(cfg),
		repo:   repo,
		cfg:    cfg,
	}
}

func (lrs *LilRagSearch) IndexNote(noteID int, content string) error {
	return lrs.IndexNoteWithNamespace(noteID, content, "")
}

func (lrs *LilRagSearch) IndexNoteWithNamespace(noteID int, content, namespace string) error {

	// Use note ID as document ID for lil-rag
	docID := fmt.Sprintf("note-%d", noteID)

	// Create namespace with ml-notes prefix
	mlNamespace := lrs.createNamespace(namespace)

	err := lrs.client.IndexDocumentWithNamespace(docID, content, mlNamespace)
	if err != nil {
		logger.Error("Failed to index note %d in lil-rag: %v", noteID, err)
		return fmt.Errorf("failed to index note in lil-rag: %w", err)
	}

	logger.Debug("Successfully indexed note %d in lil-rag", noteID)
	return nil
}

func (lrs *LilRagSearch) SearchSimilar(query string, limit int) ([]*models.Note, error) {
	return lrs.SearchSimilarWithNamespace(query, limit, "")
}

func (lrs *LilRagSearch) SearchSimilarWithNamespace(query string, limit int, namespace string) ([]*models.Note, error) {

	// Check if lil-rag is available
	if !lrs.client.IsAvailable() {
		logger.Debug("Lil-rag service not available, falling back to text search")
		return lrs.repo.Search(query)
	}

	// Create namespace with ml-notes prefix
	mlNamespace := lrs.createNamespace(namespace)

	logger.Debug("Performing lil-rag search for: %s (namespace: %s)", query, mlNamespace)

	results, err := lrs.client.SearchWithNamespace(query, limit, mlNamespace)
	if err != nil {
		logger.Error("Lil-rag search failed: %v", err)
		logger.Debug("Falling back to text search")
		return lrs.repo.Search(query)
	}

	if len(results) == 0 {
		logger.Debug("No results from lil-rag")
		return []*models.Note{}, nil
	}

	// Convert lil-rag results to notes
	var notes []*models.Note
	for _, result := range results {
		// Extract note ID from document ID (note-123 -> 123)
		noteID, err := extractNoteIDFromDocID(result.ID)
		if err != nil {
			logger.Debug("Skipping result with invalid document ID: %s", result.ID)
			continue
		}

		note, err := lrs.repo.GetByID(noteID)
		if err != nil {
			logger.Debug("Could not retrieve note %d: %v", noteID, err)
			continue
		}

		// Add score information as a temporary field if needed
		// Note: the models.Note struct doesn't have a Score field,
		// so we'll just use the notes as-is
		notes = append(notes, note)
	}

	logger.Debug("Retrieved %d notes from lil-rag search results", len(notes))
	return notes, nil
}

// IsAvailable checks if the lil-rag service is available
func (lrs *LilRagSearch) IsAvailable() bool {
	return lrs.client.IsAvailable()
}

// extractNoteIDFromDocID extracts the note ID from a lil-rag document ID
// Expected format: "note-123" -> 123
func extractNoteIDFromDocID(docID string) (int, error) {
	if len(docID) < 6 || docID[:5] != "note-" {
		return 0, fmt.Errorf("invalid document ID format: %s", docID)
	}

	noteIDStr := docID[5:]
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid note ID in document ID %s: %w", docID, err)
	}

	return noteID, nil
}

// createNamespace creates a namespace with ml-notes prefix
func (lrs *LilRagSearch) createNamespace(namespace string) string {
	if namespace == "" {
		return "ml-notes"
	}
	return fmt.Sprintf("ml-notes-%s", namespace)
}
