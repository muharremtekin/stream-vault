package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_NoOriginHeader_PassesThrough(t *testing.T) {
	called := false
	handler := CORS(DefaultCORSOptions())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no Access-Control-Allow-Origin header when no Origin sent")
	}
}

func TestCORS_WildcardOrigin_SetsStarHeader(t *testing.T) {
	handler := CORS(DefaultCORSOptions())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	acao := rec.Header().Get("Access-Control-Allow-Origin")
	if acao != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got %q", acao)
	}
}

func TestCORS_SpecificOriginAllowed_EchosOrigin(t *testing.T) {
	opts := CORSOptions{
		AllowedOrigins: []string{"http://app.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         600,
	}
	handler := CORS(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://app.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	acao := rec.Header().Get("Access-Control-Allow-Origin")
	if acao != "http://app.com" {
		t.Errorf("expected origin 'http://app.com', got %q", acao)
	}
	vary := rec.Header().Get("Vary")
	if vary != "Origin" {
		t.Errorf("expected Vary 'Origin', got %q", vary)
	}
}

func TestCORS_SpecificOriginDenied_NoHeaders(t *testing.T) {
	opts := CORSOptions{
		AllowedOrigins: []string{"http://app.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         600,
	}
	called := false
	handler := CORS(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called even for disallowed origin")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no Access-Control-Allow-Origin for disallowed origin")
	}
}

func TestCORS_PreflightOptions_Returns204(t *testing.T) {
	handler := CORS(DefaultCORSOptions())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header on preflight")
	}
	if rec.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected Access-Control-Allow-Headers header on preflight")
	}
	if rec.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("expected Access-Control-Max-Age header on preflight")
	}
}

func TestCORS_Credentials_EchosOriginNotStar(t *testing.T) {
	opts := DefaultCORSOptions()
	opts.AllowCredentials = true
	handler := CORS(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	acao := rec.Header().Get("Access-Control-Allow-Origin")
	if acao == "*" {
		t.Error("expected specific origin, not wildcard, when credentials are enabled")
	}
	if acao != "http://example.com" {
		t.Errorf("expected origin 'http://example.com', got %q", acao)
	}
	creds := rec.Header().Get("Access-Control-Allow-Credentials")
	if creds != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true', got %q", creds)
	}
}

func TestCORS_ExposedHeaders_Set(t *testing.T) {
	handler := CORS(DefaultCORSOptions())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	exposed := rec.Header().Get("Access-Control-Expose-Headers")
	if exposed == "" {
		t.Error("expected Access-Control-Expose-Headers to be set")
	}
}
