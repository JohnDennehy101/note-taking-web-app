package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotFoundResponse(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "non-existent route",
			path:           "/v1/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-existent note",
			path:           "/v1/notes/999999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response struct {
				Error string `json:"error"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			expectedMessage := "the requested resource could not be found"
			if response.Error != expectedMessage {
				t.Errorf("expected error message %q, got %q", expectedMessage, response.Error)
			}
		})
	}
}

func TestMethodNotAllowedResponse(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST to GET-only endpoint",
			method:         http.MethodPost,
			path:           "/v1/healthcheck",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET to POST-only endpoint",
			method:         http.MethodGet,
			path:           "/v1/notes",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH method not supported",
			method:         http.MethodPatch,
			path:           "/v1/notes/1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response struct {
				Error string `json:"error"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			expectedPrefix := "the " + tt.method + " method is not supported"
			if response.Error[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("expected error message to start with %q, got %q", expectedPrefix, response.Error)
			}
		})
	}
}

func TestBadRequestResponse(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			body:           `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
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

			var response struct {
				Error string `json:"error"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Error == "" {
				t.Error("expected error message to be present")
			}
		})
	}
}

func TestFailedValidationResponse(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "missing title",
			body: map[string]interface{}{
				"body": "Test body",
				"tags": []string{"test"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "missing body",
			body: map[string]interface{}{
				"title": "Test title",
				"tags":  []string{"test"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "duplicate tags",
			body: map[string]interface{}{
				"title": "Test title",
				"body":  "Test body",
				"tags":  []string{"test", "test"},
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/v1/notes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response struct {
				Error map[string]string `json:"error"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if len(response.Error) == 0 {
				t.Error("expected validation errors to be present")
			}
		})
	}
}
