package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorResponse_401(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteErrorResponse(rec, http.StatusUnauthorized, "missing token")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Status != 401 {
		t.Errorf("expected body status 401, got %d", body.Status)
	}
	if body.Error != "Unauthorized" {
		t.Errorf("expected error text 'Unauthorized', got %q", body.Error)
	}
	if body.Message != "missing token" {
		t.Errorf("expected message 'missing token', got %q", body.Message)
	}
	if body.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestWriteErrorResponse_429(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteErrorResponse(rec, http.StatusTooManyRequests, "rate limit exceeded")

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rec.Code)
	}

	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Error != "Too Many Requests" {
		t.Errorf("expected error text 'Too Many Requests', got %q", body.Error)
	}
}

func TestWriteErrorResponse_500(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteErrorResponse(rec, http.StatusInternalServerError, "something went wrong")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Status != 500 {
		t.Errorf("expected body status 500, got %d", body.Status)
	}
	if body.Error != "Internal Server Error" {
		t.Errorf("expected error text 'Internal Server Error', got %q", body.Error)
	}
	if body.Message != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got %q", body.Message)
	}
}
