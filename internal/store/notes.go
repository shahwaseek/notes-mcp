package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Note represents a single note in the database.
type Note struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Tags      string `json:"tags"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AddNote creates a new note and returns it with its generated ID and timestamps.
func (s *Store) AddNote(title, body, tags string) (*Note, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("title must not be empty")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.Exec(
		`INSERT INTO notes (title, body, tags, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		title, body, tags, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	s.logger.Info("note created", "id", id, "title", title)
	return &Note{
		ID:        id,
		Title:     title,
		Body:      body,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetNote retrieves a single note by ID.
func (s *Store) GetNote(id int64) (*Note, error) {
	n := &Note{}
	err := s.db.QueryRow(
		`SELECT id, title, body, tags, created_at, updated_at FROM notes WHERE id = ?`, id,
	).Scan(&n.ID, &n.Title, &n.Body, &n.Tags, &n.CreatedAt, &n.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("note with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("querying note: %w", err)
	}
	return n, nil
}

// SearchNotes performs a full-text search across note titles, bodies, and tags.
// The query uses SQLite FTS5 syntax. Limit defaults to 20 if <= 0.
func (s *Store) SearchNotes(query string, limit int) ([]Note, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("search query must not be empty")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT n.id, n.title, n.body, n.tags, n.created_at, n.updated_at
		FROM notes_fts fts
		JOIN notes n ON n.id = fts.rowid
		WHERE notes_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("searching notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.Tags, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning note row: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// DeleteNote removes a note by ID. Returns an error if the note does not exist.
func (s *Store) DeleteNote(id int64) error {
	result, err := s.db.Exec(`DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting note: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("note with id %d not found", id)
	}
	s.logger.Info("note deleted", "id", id)
	return nil
}

// ListRecentNotes returns the most recent notes, up to the given limit.
func (s *Store) ListRecentNotes(limit int) ([]Note, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(`
		SELECT id, title, body, tags, created_at, updated_at
		FROM notes ORDER BY id DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("listing recent notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.Tags, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning note row: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}
