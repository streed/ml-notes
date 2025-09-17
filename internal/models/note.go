package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	interrors "github.com/streed/ml-notes/internal/errors"
	"github.com/streed/ml-notes/internal/logger"
)

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(title, content string) (*Note, error) {
	result, err := r.db.Exec(
		"INSERT INTO notes (title, content) VALUES (?, ?)",
		title, content,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get insert id: %w", err)
	}

	return r.GetByID(int(id))
}

func (r *NoteRepository) GetByID(id int) (*Note, error) {
	var note Note
	err := r.db.QueryRow(
		"SELECT id, title, content, created_at, updated_at FROM notes WHERE id = ?",
		id,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, interrors.ErrNoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Load tags for this note
	tags, err := r.getTagsForNote(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	note.Tags = tags

	return &note, nil
}

func (r *NoteRepository) List(limit, offset int) ([]*Note, error) {
	// Optimized query: JOIN with tags to avoid N+1 problem
	query := `
		SELECT DISTINCT n.id, n.title, n.content, n.created_at, n.updated_at
		FROM notes n
		ORDER BY n.created_at DESC`
	args := []interface{}{}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	// First, get the notes
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Debug("Failed to close rows: %v", err)
		}
	}()

	var notes []*Note
	var noteIDs []int
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		note.Tags = []string{} // Initialize empty tags slice
		notes = append(notes, &note)
		noteIDs = append(noteIDs, note.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// If we have notes, load all their tags in a single query
	if len(noteIDs) > 0 {
		if err := r.loadTagsForNotes(notes, noteIDs); err != nil {
			return nil, fmt.Errorf("failed to load tags for notes: %w", err)
		}
	}

	return notes, nil
}

func (r *NoteRepository) UpdateByID(id int, title, content string) (*Note, error) {
	_, err := r.db.Exec(
		"UPDATE notes SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		title, content, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return r.GetByID(id)
}

func (r *NoteRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return interrors.ErrNoteNotFound
	}

	return nil
}

func (r *NoteRepository) Search(query string) ([]*Note, error) {
	return r.SearchByProject(query, "")
}

func (r *NoteRepository) SearchByProject(query, projectID string) ([]*Note, error) {
	// Try FTS first, fall back to LIKE queries if FTS is not available
	var sqlQuery string
	var args []interface{}

	// Check if FTS table exists
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='notes_fts'").Scan(&count)

	if err == nil && count > 0 {
		// Use FTS for better full-text search performance
		sqlQuery = `
			SELECT n.id, n.title, n.content, n.created_at, n.updated_at
			FROM notes n
			WHERE n.id IN (
				SELECT rowid FROM notes_fts WHERE notes_fts MATCH ?
			) OR n.title LIKE ?
			ORDER BY n.created_at DESC`
		searchQuery := "%" + query + "%"
		args = []interface{}{query, searchQuery}
	} else {
		// Fall back to LIKE queries
		searchQuery := "%" + query + "%"
		sqlQuery = "SELECT id, title, content, created_at, updated_at FROM notes WHERE (title LIKE ? OR content LIKE ?) ORDER BY created_at DESC"
		args = []interface{}{searchQuery, searchQuery}
	}

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Debug("Failed to close rows: %v", err)
		}
	}()

	var notes []*Note
	var noteIDs []int
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		note.Tags = []string{} // Initialize empty tags slice
		notes = append(notes, &note)
		noteIDs = append(noteIDs, note.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Load all tags for all notes in a single query
	if len(noteIDs) > 0 {
		if err := r.loadTagsForNotes(notes, noteIDs); err != nil {
			return nil, fmt.Errorf("failed to load tags for notes: %w", err)
		}
	}

	return notes, nil
}

// ListWithLimit returns notes with limit and offset for pagination
func (r *NoteRepository) ListWithLimit(limit, offset int) ([]*Note, error) {
	return r.List(limit, offset)
}

// Update updates a note with the given note struct
func (r *NoteRepository) Update(note *Note) error {
	_, err := r.db.Exec(
		"UPDATE notes SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		note.Title, note.Content, note.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	// Refresh the note from database to get updated timestamp
	updated, err := r.GetByID(note.ID)
	if err != nil {
		return err
	}
	*note = *updated
	return nil
}

// getTagsForNote retrieves all tags associated with a specific note
func (r *NoteRepository) getTagsForNote(noteID int) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT t.name 
		FROM tags t 
		JOIN note_tags nt ON t.id = nt.tag_id 
		WHERE nt.note_id = ? 
		ORDER BY t.name
	`, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Debug("Failed to close rows: %v", err)
		}
	}()

	var tags []string
	for rows.Next() {
		var tagName string
		if err := rows.Scan(&tagName); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tagName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}

// CreateWithTags creates a new note with associated tags
func (r *NoteRepository) CreateWithTags(title, content string, tags []string) (*Note, error) {
	return r.CreateWithTagsAndProject(title, content, tags, "default")
}

// CreateWithTagsAndProject creates a new note with associated tags in a specific project
func (r *NoteRepository) CreateWithTagsAndProject(title, content string, tags []string, projectID string) (*Note, error) {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Error("Failed to rollback transaction: %v", err)
		}
	}()

	// Create the note
	result, err := tx.Exec(
		"INSERT INTO notes (title, content) VALUES (?, ?)",
		title, content,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get insert id: %w", err)
	}

	// Add tags if provided
	if len(tags) > 0 {
		if err := r.addTagsToNoteInTx(tx, int(noteID), tags); err != nil {
			return nil, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(int(noteID))
}

// UpdateTags updates the tags associated with a note
func (r *NoteRepository) UpdateTags(noteID int, tags []string) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Error("Failed to rollback transaction: %v", err)
		}
	}()

	// Remove existing tags for this note
	_, err = tx.Exec("DELETE FROM note_tags WHERE note_id = ?", noteID)
	if err != nil {
		return fmt.Errorf("failed to remove existing tags: %w", err)
	}

	// Add new tags
	if len(tags) > 0 {
		if err := r.addTagsToNoteInTx(tx, noteID, tags); err != nil {
			return fmt.Errorf("failed to add new tags: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// addTagsToNoteInTx adds tags to a note within a transaction
func (r *NoteRepository) addTagsToNoteInTx(tx *sql.Tx, noteID int, tags []string) error {
	for _, tagName := range tags {
		if tagName == "" {
			continue // Skip empty tags
		}

		// Get or create tag
		tagID, err := r.getOrCreateTagInTx(tx, tagName)
		if err != nil {
			return fmt.Errorf("failed to get/create tag '%s': %w", tagName, err)
		}

		// Link note to tag
		_, err = tx.Exec(
			"INSERT OR IGNORE INTO note_tags (note_id, tag_id) VALUES (?, ?)",
			noteID, tagID,
		)
		if err != nil {
			return fmt.Errorf("failed to link note to tag '%s': %w", tagName, err)
		}
	}
	return nil
}

// getOrCreateTagInTx gets an existing tag or creates a new one within a transaction
func (r *NoteRepository) getOrCreateTagInTx(tx *sql.Tx, tagName string) (int, error) {
	// Try to get existing tag
	var tagID int
	err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err == nil {
		return tagID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("failed to query tag: %w", err)
	}

	// Create new tag
	result, err := tx.Exec("INSERT INTO tags (name) VALUES (?)", tagName)
	if err != nil {
		return 0, fmt.Errorf("failed to create tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get tag insert id: %w", err)
	}

	return int(id), nil
}

// SearchByTags searches for notes that have any of the specified tags
func (r *NoteRepository) SearchByTags(tags []string) ([]*Note, error) {
	return r.SearchByTagsAndProject(tags, "")
}

// SearchByTagsAndProject searches for notes that have any of the specified tags within a project
func (r *NoteRepository) SearchByTagsAndProject(tags []string, projectID string) ([]*Note, error) {
	if len(tags) == 0 {
		return []*Note{}, nil
	}

	// Build placeholders for SQL IN clause
	placeholders := make([]string, len(tags))
	args := make([]interface{}, len(tags))
	for i, tag := range tags {
		placeholders[i] = "?"
		args[i] = tag
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT n.id, n.title, n.content, n.created_at, n.updated_at 
		FROM notes n
		JOIN note_tags nt ON n.id = nt.note_id
		JOIN tags t ON nt.tag_id = t.id
		WHERE t.name IN (%s)
		ORDER BY n.created_at DESC`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes by tags: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Debug("Failed to close rows: %v", err)
		}
	}()

	var notes []*Note
	var noteIDs []int
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		note.Tags = []string{} // Initialize empty tags slice
		notes = append(notes, &note)
		noteIDs = append(noteIDs, note.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Load all tags for all notes in a single query
	if len(noteIDs) > 0 {
		if err := r.loadTagsForNotes(notes, noteIDs); err != nil {
			return nil, fmt.Errorf("failed to load tags for notes: %w", err)
		}
	}

	return notes, nil
}

// GetAllTags returns all tags in the system
func (r *NoteRepository) GetAllTags() ([]Tag, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.name, t.created_at, COUNT(nt.note_id) as note_count
		FROM tags t
		LEFT JOIN note_tags nt ON t.id = nt.tag_id
		GROUP BY t.id, t.name, t.created_at
		ORDER BY t.name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Debug("Failed to close rows: %v", err)
		}
	}()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var noteCount int
		err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &noteCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}

// GetNoteCount returns the total number of notes
func (r *NoteRepository) GetNoteCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get note count: %w", err)
	}
	return count, nil
}

// GetTagCount returns the total number of unique tags
func (r *NoteRepository) GetTagCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(DISTINCT name) FROM tags").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get tag count: %w", err)
	}
	return count, nil
}

// loadTagsForNotes efficiently loads tags for multiple notes in a single query
func (r *NoteRepository) loadTagsForNotes(notes []*Note, noteIDs []int) error {
	if len(noteIDs) == 0 {
		return nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(noteIDs))
	args := make([]interface{}, len(noteIDs))
	for i, id := range noteIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	// Single query to get all tags for all notes
	query := fmt.Sprintf(`
		SELECT nt.note_id, t.name
		FROM note_tags nt
		JOIN tags t ON nt.tag_id = t.id
		WHERE nt.note_id IN (%s)
		ORDER BY nt.note_id, t.name`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to query tags: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Debug("Failed to close rows: %v", err)
		}
	}()

	// Create a map to group tags by note ID
	tagsByNoteID := make(map[int][]string)
	for rows.Next() {
		var noteID int
		var tagName string
		if err := rows.Scan(&noteID, &tagName); err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		tagsByNoteID[noteID] = append(tagsByNoteID[noteID], tagName)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating tag rows: %w", err)
	}

	// Assign tags to notes
	for _, note := range notes {
		if tags, exists := tagsByNoteID[note.ID]; exists {
			note.Tags = tags
		} else {
			note.Tags = []string{} // Ensure non-nil slice for notes without tags
		}
	}

	return nil
}
