package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequiresAdmin(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/api/stream/upload", true},
		{"/api/stream/upload/multipart", true},
		{"/api/encoding/jobs", true},
		{"/api/encoding/jobs/abc123", true},
		{"/api/stream/continue-watching", false},
		{"/api/stream/abc/progress", false},
		{"/stream/abc/manifest.m3u8", false},
		{"/api/catalog/movies", false},
		{"/health", false},
		{"/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := requiresAdmin(tt.path)
			if got != tt.want {
				t.Errorf("requiresAdmin(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestStreamingAuth_AdminPath_AdminRole_Passes(t *testing.T) {
	called := false
	handler := StreamingAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/stream/upload", nil)
	req.Header.Set("X-User-Role", "Admin")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for Admin role on admin path")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestStreamingAuth_AdminPath_PremiumRole_Returns403(t *testing.T) {
	handler := StreamingAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for non-Admin on admin path")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/stream/upload", nil)
	req.Header.Set("X-User-Role", "Premium")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Message != "admin access required" {
		t.Errorf("unexpected message: %q", body.Message)
	}
}

func TestStreamingAuth_AdminPath_NoRole_Returns403(t *testing.T) {
	handler := StreamingAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called without role header on admin path")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/encoding/jobs", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestStreamingAuth_AdminPath_FreeRole_Returns403(t *testing.T) {
	handler := StreamingAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for Free role on admin path")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/stream/upload", nil)
	req.Header.Set("X-User-Role", "Free")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestStreamingAuth_NonAdminPath_PremiumRole_Passes(t *testing.T) {
	called := false
	handler := StreamingAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/stream/continue-watching", nil)
	req.Header.Set("X-User-Role", "Premium")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for non-admin path")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestStreamingAuth_NonAdminPath_FreeRole_Passes(t *testing.T) {
	called := false
	handler := StreamingAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/abc/manifest.m3u8", nil)
	req.Header.Set("X-User-Role", "Free")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for non-admin path even with Free role")
	}
}

func TestStreamingAuth_EncodingPath_AdminRole_Passes(t *testing.T) {
	called := false
	handler := StreamingAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/encoding/jobs/abc123", nil)
	req.Header.Set("X-User-Role", "Admin")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for Admin on encoding path")
	}
}
