package search

import (
	"database/sql"
	"fmt"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/streed/ml-notes/internal/config"
	"github.com/streed/ml-notes/internal/embeddings"
	"github.com/streed/ml-notes/internal/logger"
	"github.com/streed/ml-notes/internal/models"
)

type VectorSearch struct {
	db       *sql.DB
	repo     *models.NoteRepository
	embedder embeddings.EmbeddingProvider
	cfg      *config.Config
}

func NewVectorSearch(db *sql.DB, repo *models.NoteRepository, cfg *config.Config) *VectorSearch {
	return &VectorSearch{
		db:       db,
		repo:     repo,
		embedder: embeddings.NewLocalEmbedding(cfg),
		cfg:      cfg,
	}
}

func (vs *VectorSearch) IndexNote(noteID int, content string) error {
	if !vs.cfg.EnableVectorSearch {
		return nil // Skip indexing if vector search is disabled
	}

	// Use document type for indexing notes
	embedding, err := vs.embedder.GetEmbeddingWithType(content, embeddings.EmbeddingTypeDocument)
	if err != nil {
		return fmt.Errorf("failed to get embedding: %w", err)
	}

	embeddingBytes := embeddings.EmbeddingToBytes(embedding)

	// Store in note_embeddings table for fallback
	_, err = vs.db.Exec(
		"INSERT OR REPLACE INTO note_embeddings (note_id, embedding) VALUES (?, ?)",
		noteID, embeddingBytes,
	)
	if err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	// Try to store in vec_notes virtual table using sqlite-vec serialization
	vecBytes, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}
	
	_, err = vs.db.Exec(
		"INSERT OR REPLACE INTO vec_notes (note_id, embedding) VALUES (?, ?)",
		noteID, vecBytes,
	)
	if err != nil {
		// Log but don't fail - we can still use fallback search
		logger.Debug("Failed to index in vec_notes: %v", err)
	} else {
		logger.Debug("Indexed note %d in vec_notes", noteID)
	}

	return nil
}

func (vs *VectorSearch) SearchSimilar(query string, limit int) ([]*models.Note, error) {
	if !vs.cfg.EnableVectorSearch {
		logger.Debug("Vector search disabled, falling back to text search")
		return vs.repo.Search(query)
	}

	logger.Debug("Performing vector search for: %s", query)
	// Use search type for queries
	queryEmbedding, err := vs.embedder.GetEmbeddingWithType(query, embeddings.EmbeddingTypeSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Try vector search with vec0 extension first
	logger.Debug("Attempting vec0 search")
	notes, err := vs.searchWithVec0(queryEmbedding, limit)
	if err == nil && len(notes) > 0 {
		logger.Debug("Vec0 search successful, found %d notes", len(notes))
		return notes, nil
	}

	if err != nil {
		logger.Debug("Vec0 search failed: %v", err)
	}

	// Fallback to manual cosine similarity
	logger.Debug("Falling back to manual cosine similarity search")
	return vs.searchWithCosineSimilarity(queryEmbedding, limit)
}

func (vs *VectorSearch) searchWithVec0(queryEmbedding []float32, limit int) ([]*models.Note, error) {
	// Serialize query embedding using sqlite-vec
	queryBytes, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query embedding: %w", err)
	}
	
	// Use vec_distance_L2 for similarity search
	rows, err := vs.db.Query(`
		SELECT note_id, vec_distance_L2(embedding, ?) as distance
		FROM vec_notes 
		ORDER BY distance 
		LIMIT ?
	`, queryBytes, limit)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var noteIDs []int
	for rows.Next() {
		var id int
		var distance float64
		if err := rows.Scan(&id, &distance); err != nil {
			continue
		}
		noteIDs = append(noteIDs, id)
	}

	var notes []*models.Note
	for _, id := range noteIDs {
		note, err := vs.repo.GetByID(id)
		if err == nil {
			notes = append(notes, note)
		}
	}

	return notes, nil
}

func (vs *VectorSearch) searchWithCosineSimilarity(queryEmbedding []float32, limit int) ([]*models.Note, error) {
	rows, err := vs.db.Query("SELECT note_id, embedding FROM note_embeddings")
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}
	defer rows.Close()

	type similarity struct {
		noteID int
		score  float32
	}

	var similarities []similarity

	for rows.Next() {
		var noteID int
		var embeddingBytes []byte

		if err := rows.Scan(&noteID, &embeddingBytes); err != nil {
			continue
		}

		noteEmbedding, err := embeddings.BytesToEmbedding(embeddingBytes)
		if err != nil {
			continue
		}

		score := embeddings.CosineSimilarity(queryEmbedding, noteEmbedding)
		similarities = append(similarities, similarity{noteID: noteID, score: score})
	}

	// Sort by similarity score
	for i := 0; i < len(similarities); i++ {
		for j := i + 1; j < len(similarities); j++ {
			if similarities[j].score > similarities[i].score {
				similarities[i], similarities[j] = similarities[j], similarities[i]
			}
		}
	}

	// Get top N notes
	var notes []*models.Note
	for i := 0; i < limit && i < len(similarities); i++ {
		note, err := vs.repo.GetByID(similarities[i].noteID)
		if err == nil {
			notes = append(notes, note)
		}
	}

	return notes, nil
}