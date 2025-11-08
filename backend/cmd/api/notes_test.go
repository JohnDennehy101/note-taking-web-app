package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johndennehy101/note-taking-web-app/backend/internal/data"
)

func TestCreateNoteHandler(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid note",
			body: map[string]interface{}{
				"title": "Test Note",
				"body":  "This is a test note body",
				"tags":  []string{"test", "example"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing title",
			body: map[string]interface{}{
				"body": "This is a test note body",
				"tags": []string{"test"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "missing body",
			body: map[string]interface{}{
				"title": "Test Note",
				"tags":  []string{"test"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "missing tags",
			body: map[string]interface{}{
				"title": "Test Note",
				"body":  "This is a test note body",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "duplicate tags",
			body: map[string]interface{}{
				"title": "Test Note",
				"body":  "This is a test note body",
				"tags":  []string{"test", "test"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "title too long",
			body: map[string]interface{}{
				"title": string(make([]byte, 501)),
				"body":  "This is a test note body",
				"tags":  []string{"test"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "invalid JSON",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				body = []byte("{ invalid json }")
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/notes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response struct {
					Note data.Note `json:"note"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Note.ID == 0 {
					t.Error("expected note ID to be set")
				}
				if response.Note.Title != tt.body["title"] {
					t.Errorf("expected title %v, got %s", tt.body["title"], response.Note.Title)
				}
			}
		})
	}
}

func TestShowNoteHandler(t *testing.T) {
	app := newTestApplication(t)

	note := createTestNote(t, app, "Test Note", "Test Body", []string{"test"})

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id",
			id:             fmt.Sprintf("%d", note.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent id",
			id:             "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			id:             "0",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "negative id",
			id:             "-1",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/notes/"+tt.id, http.NoBody)
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Note data.Note `json:"note"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Note.ID == 0 {
					t.Error("expected note ID to be set")
				}
				if response.Note.ID != note.ID {
					t.Errorf("expected note ID %d, got %d", note.ID, response.Note.ID)
				}
			}
		})
	}
}

func TestUpdateNoteHandler(t *testing.T) {
	app := newTestApplication(t)

	note := createTestNote(t, app, "Original Title", "Original Body", []string{"original"})

	tests := []struct {
		name           string
		id             int64
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid update",
			id:   note.ID,
			body: map[string]interface{}{
				"title": "Updated Title",
				"body":  "Updated Body",
				"tags":  []string{"updated"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "non-existent id",
			id:   999,
			body: map[string]interface{}{
				"title": "Updated Title",
				"body":  "Updated Body",
				"tags":  []string{"updated"},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "missing title",
			id:   note.ID,
			body: map[string]interface{}{
				"body": "Updated Body",
				"tags": []string{"updated"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "archive note",
			id:   note.ID,
			body: map[string]interface{}{
				"title":    "Archived Note",
				"body":     "This note is archived",
				"tags":     []string{"archived"},
				"archived": true,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPut, "/v1/notes/"+fmt.Sprintf("%d", tt.id), bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Note data.Note `json:"note"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Note.Title != tt.body["title"] {
					t.Errorf("expected title %v, got %s", tt.body["title"], response.Note.Title)
				}
			}
		})
	}
}

func TestDeleteNoteHandler(t *testing.T) {
	app := newTestApplication(t)

	note := createTestNote(t, app, "To Delete", "This will be deleted", []string{"delete"})

	tests := []struct {
		name           string
		id             int64
		expectedStatus int
	}{
		{
			name:           "valid delete",
			id:             note.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent id",
			id:             999,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/v1/notes/"+fmt.Sprintf("%d", tt.id), http.NoBody)
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Message string `json:"message"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Message != "note successfully deleted" {
					t.Errorf("expected message 'note successfully deleted', got %s", response.Message)
				}
			}
		})
	}
}

func TestCreateNoteDefaultsArchivedToFalse(t *testing.T) {
	app := newTestApplication(t)

	body := map[string]interface{}{
		"title": "New Note",
		"body":  "Note body",
		"tags":  []string{"test"},
	}

	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/notes", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := app.routes()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var response struct {
		Note data.Note `json:"note"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Note.Archived != false {
		t.Errorf("expected archived to be false by default, got %v", response.Note.Archived)
	}
}

func TestArchivedPropertyUpdate(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		setupArchived  bool
		updateArchived bool
		expectedStatus int
		verifyArchived bool
		expectedValue  bool
	}{
		{
			name:           "archive unarchived note",
			setupArchived:  false,
			updateArchived: true,
			expectedStatus: http.StatusOK,
			verifyArchived: true,
			expectedValue:  true,
		},
		{
			name:           "unarchive archived note",
			setupArchived:  true,
			updateArchived: false,
			expectedStatus: http.StatusOK,
			verifyArchived: true,
			expectedValue:  false,
		},
		{
			name:           "keep note archived",
			setupArchived:  true,
			updateArchived: true,
			expectedStatus: http.StatusOK,
			verifyArchived: true,
			expectedValue:  true,
		},
		{
			name:           "keep note unarchived",
			setupArchived:  false,
			updateArchived: false,
			expectedStatus: http.StatusOK,
			verifyArchived: true,
			expectedValue:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := createTestNote(t, app, "Test Note", "Test Body", []string{"test"})

			if tt.setupArchived {
				archiveBody := map[string]interface{}{
					"title":    note.Title,
					"body":     note.Body,
					"tags":     note.Tags,
					"archived": true,
				}
				reqBody, _ := json.Marshal(archiveBody)
				req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/notes/%d", note.ID), bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				router := app.routes()
				router.ServeHTTP(rr, req)
			}

			body := map[string]interface{}{
				"title":    note.Title,
				"body":     note.Body,
				"tags":     note.Tags,
				"archived": tt.updateArchived,
			}

			reqBody, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/notes/%d", note.ID), bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.verifyArchived {
				var response struct {
					Note data.Note `json:"note"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
				if response.Note.Archived != tt.expectedValue {
					t.Errorf("expected archived to be %v, got %v", tt.expectedValue, response.Note.Archived)
				}
			}
		})
	}
}

func TestArchivedPropertyGet(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		archived       bool
		expectedStatus int
	}{
		{
			name:           "show archived note",
			archived:       true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "show unarchived note",
			archived:       false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := createTestNote(t, app, "Test Note", "Body", []string{"test"})

			if tt.archived {
				archiveBody := map[string]interface{}{
					"title":    note.Title,
					"body":     note.Body,
					"tags":     note.Tags,
					"archived": true,
				}
				reqBody, _ := json.Marshal(archiveBody)
				req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/notes/%d", note.ID), bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				router := app.routes()
				router.ServeHTTP(rr, req)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/notes/%d", note.ID), http.NoBody)
			rr := httptest.NewRecorder()
			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response struct {
				Note data.Note `json:"note"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}

			if response.Note.Archived != tt.archived {
				t.Errorf("expected archived to be %v, got %v", tt.archived, response.Note.Archived)
			}
		})
	}
}
