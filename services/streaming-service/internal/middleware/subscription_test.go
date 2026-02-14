package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubscription_FreeRole_Returns403(t *testing.T) {
	handler := Subscription()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for Free role")
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.Header.Set("X-User-Role", "Free")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Message != "subscription required for streaming" {
		t.Errorf("unexpected message: %q", body.Message)
	}
}

func TestSubscription_ProtoFreeRole_Returns403(t *testing.T) {
	handler := Subscription()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for SUBSCRIPTION_TIER_FREE role")
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.Header.Set("X-User-Role", "SUBSCRIPTION_TIER_FREE")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestSubscription_EmptyRole_DefaultsToFree_Returns403(t *testing.T) {
	handler := Subscription()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for empty role (defaults to Free)")
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	// No X-User-Role header set
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestSubscription_BasicRole_Passes(t *testing.T) {
	called := false
	handler := Subscription()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.Header.Set("X-User-Role", "Basic")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for Basic role")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestSubscription_StandardRole_Passes(t *testing.T) {
	called := false
	handler := Subscription()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.Header.Set("X-User-Role", "Standard")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for Standard role")
	}
}

func TestSubscription_PremiumRole_Passes(t *testing.T) {
	called := false
	handler := Subscription()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.Header.Set("X-User-Role", "Premium")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for Premium role")
	}
}

func TestSubscription_AdminRole_Passes(t *testing.T) {
	called := false
	handler := Subscription()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.Header.Set("X-User-Role", "Admin")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for Admin role")
	}
}
