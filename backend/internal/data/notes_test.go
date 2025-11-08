package data_test

import (
	"testing"

	"github.com/johndennehy101/note-taking-web-app/backend/internal/data"
	"github.com/johndennehy101/note-taking-web-app/backend/internal/testutil"
	"github.com/johndennehy101/note-taking-web-app/backend/internal/validator"
)

func newTestModel(t *testing.T) data.NoteModel {
	db, err := testutil.GetTestDB()
	if err != nil {
		t.Fatalf("failed to get test DB: %v", err)
	}
	return data.NoteModel{DB: db}
}

func TestNoteModel_Insert(t *testing.T) {
	model := newTestModel(t)

	tests := []struct {
		name    string
		note    *data.Note
		wantErr bool
	}{
		{
			name: "valid note",
			note: &data.Note{
				Title: "Test Note",
				Body:  "Test Body",
				Tags:  []string{"test", "example"},
			},
			wantErr: false,
		},
		{
			name: "note with single tag",
			note: &data.Note{
				Title: "Single Tag Note",
				Body:  "Body",
				Tags:  []string{"single"},
			},
			wantErr: false,
		},
		{
			name: "note with empty tags array",
			note: &data.Note{
				Title: "Empty Tags",
				Body:  "Body",
				Tags:  []string{},
			},
			wantErr: false,
		},
		{
			name: "note with many tags",
			note: &data.Note{
				Title: "Many Tags",
				Body:  "Body",
				Tags:  []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.Insert(tt.note)
			if (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if tt.note.ID == 0 {
					t.Error("expected ID to be set")
				}
				if tt.note.CreatedAt.IsZero() {
					t.Error("expected CreatedAt to be set")
				}
				if tt.note.UpdatedAt.IsZero() {
					t.Error("expected UpdatedAt to be set")
				}
				if tt.note.Version != 1 {
					t.Errorf("expected Version to be 1, got %d", tt.note.Version)
				}
				if tt.note.Archived != false {
					t.Errorf("expected Archived to be false by default, got %v", tt.note.Archived)
				}
			}
		})
	}
}

func TestNoteModel_Get(t *testing.T) {
	model := newTestModel(t)

	note := &data.Note{
		Title: "Test Note",
		Body:  "Test Body",
		Tags:  []string{"test"},
	}
	err := model.Insert(note)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		wantID  int64
	}{
		{
			name:    "valid id",
			id:      note.ID,
			wantErr: false,
			wantID:  note.ID,
		},
		{
			name:    "non-existent id",
			id:      999999,
			wantErr: true,
		},
		{
			name:    "invalid id - zero",
			id:      0,
			wantErr: true,
		},
		{
			name:    "invalid id - negative",
			id:      -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.Get(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if got.ID != tt.wantID {
					t.Errorf("expected ID %d, got %d", tt.wantID, got.ID)
				}
				if got.Title != note.Title {
					t.Errorf("expected Title %s, got %s", note.Title, got.Title)
				}
				if got.Body != note.Body {
					t.Errorf("expected Body %s, got %s", note.Body, got.Body)
				}
				if len(got.Tags) != len(note.Tags) {
					t.Errorf("expected %d tags, got %d", len(note.Tags), len(got.Tags))
				}
			} else if err != data.ErrRecordNotFound {
				t.Errorf("expected ErrRecordNotFound, got %v", err)
			}

		})
	}
}

func TestNoteModel_Update(t *testing.T) {
	model := newTestModel(t)

	note := &data.Note{
		Title: "Original Title",
		Body:  "Original Body",
		Tags:  []string{"original"},
	}
	err := model.Insert(note)
	if err != nil {
		t.Fatal(err)
	}

	originalVersion := note.Version

	tests := []struct {
		name             string
		updateNote       *data.Note
		wantErr          bool
		expectedTitle    string
		expectedBody     string
		expectedTags     []string
		expectedArchived bool
		versionIncrement bool
	}{
		{
			name: "update all fields",
			updateNote: &data.Note{
				ID:       note.ID,
				Title:    "Updated Title",
				Body:     "Updated Body",
				Tags:     []string{"updated"},
				Archived: true,
			},
			wantErr:          false,
			expectedTitle:    "Updated Title",
			expectedBody:     "Updated Body",
			expectedTags:     []string{"updated"},
			expectedArchived: true,
			versionIncrement: true,
		},
		{
			name: "update without archiving",
			updateNote: &data.Note{
				ID:       note.ID,
				Title:    "New Title",
				Body:     "New Body",
				Tags:     []string{"new"},
				Archived: false,
			},
			wantErr:          false,
			expectedTitle:    "New Title",
			expectedBody:     "New Body",
			expectedTags:     []string{"new"},
			expectedArchived: false,
			versionIncrement: true,
		},
		{
			name: "update tags only",
			updateNote: &data.Note{
				ID:       note.ID,
				Title:    "New Title",
				Body:     "New Body",
				Tags:     []string{"tag1", "tag2"},
				Archived: false,
			},
			wantErr:          false,
			expectedTitle:    "New Title",
			expectedBody:     "New Body",
			expectedTags:     []string{"tag1", "tag2"},
			expectedArchived: false,
			versionIncrement: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.Update(tt.updateNote)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if tt.versionIncrement {
					expectedVersion := originalVersion + 1
					if tt.updateNote.Version != expectedVersion {
						t.Errorf("expected version to increment from %d to %d, got %d",
							originalVersion, expectedVersion, tt.updateNote.Version)
					}
					originalVersion = tt.updateNote.Version
				}

				updated, err := model.Get(tt.updateNote.ID)
				if err != nil {
					t.Fatal(err)
				}

				if updated.Title != tt.expectedTitle {
					t.Errorf("expected Title %s, got %s", tt.expectedTitle, updated.Title)
				}
				if updated.Body != tt.expectedBody {
					t.Errorf("expected Body %s, got %s", tt.expectedBody, updated.Body)
				}
				if len(updated.Tags) != len(tt.expectedTags) {
					t.Errorf("expected %d tags, got %d", len(tt.expectedTags), len(updated.Tags))
				}
				if updated.Archived != tt.expectedArchived {
					t.Errorf("expected Archived %v, got %v", tt.expectedArchived, updated.Archived)
				}
			}
		})
	}
}

func TestNoteModel_Delete(t *testing.T) {
	model := newTestModel(t)

	tests := []struct {
		name    string
		setup   func() int64
		wantErr bool
	}{
		{
			name: "valid delete",
			setup: func() int64 {
				note := &data.Note{
					Title: "To Delete",
					Body:  "This will be deleted",
					Tags:  []string{"delete"},
				}
				_ = model.Insert(note)
				return note.ID
			},
			wantErr: false,
		},
		{
			name: "non-existent id",
			setup: func() int64 {
				return 999999
			},
			wantErr: true,
		},
		{
			name: "invalid id - zero",
			setup: func() int64 {
				return 0
			},
			wantErr: true,
		},
		{
			name: "invalid id - negative",
			setup: func() int64 {
				return -1
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setup()
			err := model.Delete(id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				_, err := model.Get(id)
				if err != data.ErrRecordNotFound {
					t.Errorf("expected ErrRecordNotFound after deletion, got %v", err)
				}
			} else if err != data.ErrRecordNotFound {
				t.Errorf("expected ErrRecordNotFound, got %v", err)
			}
		})
	}
}

func TestValidateNote(t *testing.T) {
	tests := []struct {
		name           string
		note           *data.Note
		expectedValid  bool
		expectedErrors []string
	}{
		{
			name: "valid note",
			note: &data.Note{
				Title: "Valid Title",
				Body:  "Valid Body",
				Tags:  []string{"tag1", "tag2"},
			},
			expectedValid:  true,
			expectedErrors: []string{},
		},
		{
			name: "missing title",
			note: &data.Note{
				Title: "",
				Body:  "Valid Body",
				Tags:  []string{"tag1"},
			},
			expectedValid:  false,
			expectedErrors: []string{"title"},
		},
		{
			name: "missing body",
			note: &data.Note{
				Title: "Valid Title",
				Body:  "",
				Tags:  []string{"tag1"},
			},
			expectedValid:  false,
			expectedErrors: []string{"body"},
		},
		{
			name: "nil tags",
			note: &data.Note{
				Title: "Valid Title",
				Body:  "Valid Body",
				Tags:  nil,
			},
			expectedValid:  false,
			expectedErrors: []string{"tags"},
		},
		{
			name: "duplicate tags",
			note: &data.Note{
				Title: "Valid Title",
				Body:  "Valid Body",
				Tags:  []string{"tag1", "tag1"},
			},
			expectedValid:  false,
			expectedErrors: []string{"tags"},
		},
		{
			name: "title too long",
			note: &data.Note{
				Title: string(make([]byte, 501)),
				Body:  "Valid Body",
				Tags:  []string{"tag1"},
			},
			expectedValid:  false,
			expectedErrors: []string{"title"},
		},
		{
			name: "title at max length",
			note: &data.Note{
				Title: string(make([]byte, 500)),
				Body:  "Valid Body",
				Tags:  []string{"tag1"},
			},
			expectedValid:  true,
			expectedErrors: []string{},
		},
		{
			name: "multiple validation errors",
			note: &data.Note{
				Title: "",
				Body:  "",
				Tags:  nil,
			},
			expectedValid:  false,
			expectedErrors: []string{"title", "body", "tags"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			data.ValidateNote(v, tt.note)

			if v.Valid() != tt.expectedValid {
				t.Errorf("expected valid=%v, got %v", tt.expectedValid, v.Valid())
			}

			if !tt.expectedValid {
				for _, expectedError := range tt.expectedErrors {
					if _, exists := v.Errors[expectedError]; !exists {
						t.Errorf("expected error for field %s, but it was not found", expectedError)
					}
				}
			}
		})
	}
}
