package main_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheckHandler(t *testing.T) {
	app := newTestApplication(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", http.NoBody)
	rr := httptest.NewRecorder()

	router := app.routes()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %s", rr.Header().Get("Content-Type"))
	}

	var response struct {
		Status     string            `json:"status"`
		SystemInfo map[string]string `json:"system_info"`
	}

	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "available" {
		t.Errorf("expected status 'available', got %s", response.Status)
	}

	if response.SystemInfo == nil {
		t.Fatal("expected system_info to be present")
	}

	if response.SystemInfo["environment"] != "testing" {
		t.Errorf("expected environment 'testing', got %s", response.SystemInfo["environment"])
	}

	if response.SystemInfo["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", response.SystemInfo["version"])
	}
}
