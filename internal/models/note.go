package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	
	interrors "github.com/streed/ml-notes/internal/errors"
)

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
