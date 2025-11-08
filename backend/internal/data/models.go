package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Notes NoteModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Notes: NoteModel{DB: db},
	}
}
