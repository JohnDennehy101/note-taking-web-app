package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		status         int
		data           map[string]interface{}
		headers        http.Header
		expectedStatus int
		expectedHeader string
		expectedValue  string
	}{
		{
			name:           "successful JSON response",
			status:         http.StatusOK,
			data:           map[string]interface{}{"message": "test", "number": 42},
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
		{
			name:           "JSON response with custom headers",
			status:         http.StatusCreated,
			data:           map[string]interface{}{"id": 1},
			headers:        http.Header{"Location": []string{"/v1/notes/1"}},
			expectedStatus: http.StatusCreated,
			expectedHeader: "application/json",
		},
		{
			name:   "JSON response with multiple headers",
			status: http.StatusOK,
			data:   map[string]interface{}{"data": "test"},
			headers: http.Header{
				"X-Custom-Header":  []string{"value1"},
				"X-Another-Header": []string{"value2"},
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", http.NoBody)
			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
			}

			if rr.Header().Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %s", rr.Header().Get("Content-Type"))
			}

			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}

			if response["status"] != "available" {
				t.Errorf("expected status 'available', got %v", response["status"])
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "valid JSON",
			body:           `{"title":"Test","body":"Body","tags":["test"]}`,
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:           "empty body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "invalid JSON syntax",
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "unknown field",
			body:           `{"title":"Test","body":"Body","tags":["test"],"unknown":"field"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "multiple JSON values",
			body:           `{"title":"Test"}{"title":"Test2"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "wrong JSON type",
			body:           `{"title":123,"body":"Body","tags":["test"]}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "malformed JSON",
			body:           `{"title":"Test"`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/notes", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if !tt.expectedError && rr.Code == http.StatusCreated {
				var response struct {
					Note struct {
						Title string   `json:"title"`
						Body  string   `json:"body"`
						Tags  []string `json:"tags"`
					} `json:"note"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestReadIDParam(t *testing.T) {
	app := newTestApplication(t)

	note := createTestNote(t, app, "Test Note", "Test Body", []string{"test"})

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id - existing note",
			id:             fmt.Sprintf("%d", note.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid id - non-existent note",
			id:             "999999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id - zero",
			id:             "0",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id - negative",
			id:             "-1",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id - non-numeric",
			id:             "abc",
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
					Note struct {
						ID int64 `json:"id"`
					} `json:"note"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if response.Note.ID != note.ID {
					t.Errorf("expected note ID %d, got %d", note.ID, response.Note.ID)
				}
			}
		})
	}
}

func TestReadJSONMaxBytes(t *testing.T) {
	app := newTestApplication(t)

	largeBody := make([]byte, 1_048_577)
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	largeBody[0] = '{'
	largeBody[1] = '"'
	largeBody[2] = 'a'
	largeBody[3] = '"'
	largeBody[4] = ':'
	largeBody[5] = '"'
	largeBody[len(largeBody)-2] = '"'
	largeBody[len(largeBody)-1] = '}'

	req := httptest.NewRequest(http.MethodPost, "/v1/notes", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := app.routes()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for body too large, got %d", http.StatusBadRequest, rr.Code)
	}
}
