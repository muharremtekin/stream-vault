package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/streamvault/streaming-service/internal/config"
)

func TestUpload_EmptyUserID_Returns401(t *testing.T) {
	h := NewUploadHandler(
		&mockStorage{},
		&mockPublisher{},
		config.UploadConfig{AllowedTypes: []string{"video/mp4"}},
		config.MinIOConfig{RawBucket: "raw"},
	)

	// No X-User-Id header set
	req := httptest.NewRequest(http.MethodPost, "/api/streaming/upload", strings.NewReader("dummy"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=testboundary")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for empty X-User-Id, got %d", rr.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if !strings.Contains(resp.Message, "user identification") {
		t.Errorf("expected error about user identification, got: %s", resp.Message)
	}
}

func TestUpload_WithUserID_PassesValidation(t *testing.T) {
	h := NewUploadHandler(
		&mockStorage{},
		&mockPublisher{},
		config.UploadConfig{AllowedTypes: []string{"video/mp4"}},
		config.MinIOConfig{RawBucket: "raw"},
	)

	// With X-User-Id but invalid body (not real multipart) — should fail on multipart parsing, not on user ID
	req := httptest.NewRequest(http.MethodPost, "/api/streaming/upload", strings.NewReader("not-multipart"))
	req.Header.Set("X-User-Id", "admin-1")
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// Should NOT be 401 — the user ID validation passed
	if rr.Code == http.StatusUnauthorized {
		t.Error("unexpected 401 — X-User-Id was provided but still rejected")
	}
	// Should fail with 400 (multipart form data required)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for non-multipart body, got %d", rr.Code)
	}
}
