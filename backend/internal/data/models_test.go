package data_test

import (
	"database/sql"
	"testing"

	"github.com/johndennehy101/note-taking-web-app/backend/internal/data"
	"github.com/johndennehy101/note-taking-web-app/backend/internal/testutil"
)

func TestNewModels(t *testing.T) {
	db, err := testutil.GetTestDB()
	if err != nil {
		t.Fatalf("failed to get test DB: %v", err)
	}

	tests := []struct {
		name string
		db   *sql.DB
	}{
		{
			name: "valid database connection",
			db:   db,
		},
		{
			name: "nil database",
			db:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := data.NewModels(tt.db)

			if models.Notes.DB != tt.db {
				t.Errorf("expected Notes.DB to be set to provided db")
			}

			// Verify Notes is properly initialized
			if models.Notes.DB == nil && tt.db != nil {
				t.Error("expected Notes.DB to be set when db is provided")
			}
		})
	}
}

func TestErrRecordNotFound(t *testing.T) {
	// Test that ErrRecordNotFound is properly exported and usable
	if data.ErrRecordNotFound == nil {
		t.Error("expected ErrRecordNotFound to be defined")
	}

	expectedMsg := "record not found"
	if data.ErrRecordNotFound.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, data.ErrRecordNotFound.Error())
	}
}
