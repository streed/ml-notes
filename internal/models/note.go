package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	interrors "github.com/streed/ml-notes/internal/errors"
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
	query := "SELECT id, title, content, created_at, updated_at FROM notes ORDER BY created_at DESC"
	args := []interface{}{}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		
		// Load tags for this note
		tags, err := r.getTagsForNote(note.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags for note %d: %w", note.ID, err)
		}
		note.Tags = tags
		
		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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
	searchQuery := "%" + query + "%"
	rows, err := r.db.Query(
		"SELECT id, title, content, created_at, updated_at FROM notes WHERE title LIKE ? OR content LIKE ? ORDER BY created_at DESC",
		searchQuery, searchQuery,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		
		// Load tags for this note
		tags, err := r.getTagsForNote(note.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags for note %d: %w", note.ID, err)
		}
		note.Tags = tags
		
		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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
	defer rows.Close()

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
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

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
	defer tx.Rollback()

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
		ORDER BY n.created_at DESC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes by tags: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		// Load tags for this note
		noteTags, err := r.getTagsForNote(note.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags for note %d: %w", note.ID, err)
		}
		note.Tags = noteTags

		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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
	defer rows.Close()

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
