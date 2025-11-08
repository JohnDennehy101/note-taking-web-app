package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/johndennehy101/note-taking-web-app/backend/internal/validator"
	"github.com/lib/pq"
)

type NoteModel struct {
	DB *sql.DB
}

type Note struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tags"`
	Archived  bool      `json:"archived"`
	Version   int       `json:"version"`
}

func (m NoteModel) Insert(note *Note) error {
	query := `
        INSERT INTO notes (title, body, tags) 
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at, version, archived`

	args := []any{note.Title, note.Body, pq.Array(note.Tags)}

	return m.DB.QueryRow(query, args...).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt, &note.Version, &note.Archived)
}

func (m NoteModel) Get(id int64) (*Note, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, created_at, updated_at, title, body, tags, archived, version
        FROM notes
        WHERE id = $1`

	var note Note

	err := m.DB.QueryRow(query, id).Scan(
		&note.ID,
		&note.CreatedAt,
		&note.UpdatedAt,
		&note.Title,
		&note.Body,
		pq.Array(&note.Tags),
		&note.Archived,
		&note.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &note, nil
}

func (m NoteModel) Update(note *Note) error {
	query := `
        UPDATE notes
        SET title = $1, body = $2, tags = $3, archived = $4, version = version + 1
        WHERE id = $5
        RETURNING version`

	args := []any{
		note.Title,
		note.Body,
		pq.Array(note.Tags),
		note.Archived,
		note.ID,
	}

	return m.DB.QueryRow(query, args...).Scan(&note.Version)
}

func (m NoteModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM notes
        WHERE id = $1`

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateNote(v *validator.Validator, note *Note) {
	v.Check(note.Title != "", "title", "must be provided")
	v.Check(len(note.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(note.Body != "", "body", "must be provided")

	v.Check(note.Tags != nil, "tags", "must be provided")

	v.Check(validator.Unique(note.Tags), "tags", "must not contain duplicate values")
}
