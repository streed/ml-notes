package search

import (
	"fmt"
	"strconv"
	"strings"

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
	return lrs.IndexNoteWithNamespace(noteID, content, "", "default")
}

func (lrs *LilRagSearch) IndexNoteWithNamespace(noteID int, content, namespace, projectID string) error {

	// Use project-specific note ID as document ID for lil-rag
	docID := fmt.Sprintf("notes-%s-%d", projectID, noteID)

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
	return lrs.SearchSimilarWithNamespace(query, limit, "", "default")
}

func (lrs *LilRagSearch) SearchSimilarWithNamespace(query string, limit int, namespace, projectID string) ([]*models.Note, error) {

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
		// Extract note ID from document ID (notes-project-123 -> 123)
		noteID, extractedProjectID, err := extractNoteIDFromDocID(result.ID)
		if err != nil {
			logger.Debug("Skipping result with invalid document ID: %s", result.ID)
			continue
		}

		// Skip results that don't match the requested project
		if extractedProjectID != projectID {
			logger.Debug("Skipping result from different project: %s vs %s", extractedProjectID, projectID)
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

func (lrs *LilRagSearch) DeleteNote(noteID int) error {
	return lrs.DeleteNoteWithNamespace(noteID, "", "default")
}

func (lrs *LilRagSearch) DeleteNoteWithNamespace(noteID int, namespace, projectID string) error {
	// Use project-specific note ID as document ID for lil-rag
	docID := fmt.Sprintf("notes-%s-%d", projectID, noteID)

	// Create namespace with ml-notes prefix
	mlNamespace := lrs.createNamespace(namespace)

	err := lrs.client.DeleteDocumentWithNamespace(docID, mlNamespace)
	if err != nil {
		logger.Error("Failed to delete note %d from lil-rag: %v", noteID, err)
		return fmt.Errorf("failed to delete note from lil-rag: %w", err)
	}

	logger.Debug("Successfully deleted note %d from lil-rag", noteID)
	return nil
}

// extractNoteIDFromDocID extracts the note ID and project ID from a lil-rag document ID
// Expected format: "notes-project-123" -> (123, "project")
func extractNoteIDFromDocID(docID string) (int, string, error) {
	if len(docID) < 8 || docID[:6] != "notes-" {
		return 0, "", fmt.Errorf("invalid document ID format: %s", docID)
	}

	parts := strings.Split(docID, "-")
	if len(parts) < 3 {
		return 0, "", fmt.Errorf("invalid document ID format: %s", docID)
	}

	// Reconstruct project ID (everything between "notes-" and the last "-")
	projectID := strings.Join(parts[1:len(parts)-1], "-")
	noteIDStr := parts[len(parts)-1]

	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid note ID in document ID %s: %w", docID, err)
	}

	return noteID, projectID, nil
}

// createNamespace creates a namespace with ml-notes prefix
func (lrs *LilRagSearch) createNamespace(namespace string) string {
	if namespace == "" {
		return "ml-notes"
	}
	return fmt.Sprintf("ml-notes-%s", namespace)
}
