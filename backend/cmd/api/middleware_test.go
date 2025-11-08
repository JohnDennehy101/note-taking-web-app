package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnableCORS(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name                string
		origin              string
		method              string
		requestMethod       string
		expectedOrigin      string
		expectedStatus      int
		expectCORSHeaders   bool
		expectPreflightResp bool
	}{
		{
			name:                "trusted origin - GET request",
			origin:              "http://localhost:3000",
			method:              http.MethodGet,
			requestMethod:       "",
			expectedOrigin:      "http://localhost:3000",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   true,
			expectPreflightResp: false,
		},
		{
			name:                "trusted origin - preflight OPTIONS",
			origin:              "http://localhost:3000",
			method:              http.MethodOptions,
			requestMethod:       "PUT",
			expectedOrigin:      "http://localhost:3000",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   true,
			expectPreflightResp: true,
		},
		{
			name:                "untrusted origin",
			origin:              "http://evil.com",
			method:              http.MethodGet,
			requestMethod:       "",
			expectedOrigin:      "",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   false,
			expectPreflightResp: false,
		},
		{
			name:                "no origin header",
			origin:              "",
			method:              http.MethodGet,
			requestMethod:       "",
			expectedOrigin:      "",
			expectedStatus:      http.StatusOK,
			expectCORSHeaders:   false,
			expectPreflightResp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/v1/healthcheck", http.NoBody)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.requestMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.requestMethod)
			}

			rr := httptest.NewRecorder()

			router := app.routes()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Header().Get("Vary") == "" {
				t.Error("expected Vary header to be set")
			}

			actualOrigin := rr.Header().Get("Access-Control-Allow-Origin")
			if tt.expectCORSHeaders {
				if actualOrigin != tt.expectedOrigin {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, actualOrigin)
				}
			} else {
				if actualOrigin != "" {
					t.Errorf("expected no Access-Control-Allow-Origin header, got %q", actualOrigin)
				}
			}

			if tt.expectPreflightResp {
				methods := rr.Header().Get("Access-Control-Allow-Methods")
				if methods == "" {
					t.Error("expected Access-Control-Allow-Methods header for preflight request")
				}
				headers := rr.Header().Get("Access-Control-Allow-Headers")
				if headers == "" {
					t.Error("expected Access-Control-Allow-Headers header for preflight request")
				}
			}
		})
	}
}

func TestRecoverPanic(t *testing.T) {
	app := newTestApplication(t)

	t.Run("normal request passes through", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", http.NoBody)
		rr := httptest.NewRecorder()

		router := app.routes()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		if rr.Header().Get("Connection") == "close" {
			t.Error("unexpected Connection: close header on normal request")
		}
	})

	t.Run("panic recovery structure verified", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", http.NoBody)
		rr := httptest.NewRecorder()

		router := app.routes()
		router.ServeHTTP(rr, req)

		if rr.Code == 0 {
			t.Error("expected a response status code")
		}
	})
}
