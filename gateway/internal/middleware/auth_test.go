package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret = "StreamVault-Test-Secret-Key-Must-Be-At-Least-32-Characters!!"
	testIssuer = "StreamVault.Test"
)

func generateTestJWT(secret, issuer, subject, role string, expiresAt time.Time) string {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuth_PublicRoute_ApiAuth_BypassesAuth(t *testing.T) {
	called := false
	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for public route /api/auth/login")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuth_PublicRoute_ApiCatalog_BypassesAuth(t *testing.T) {
	called := false
	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/catalog/movies", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for public route /api/catalog/movies")
	}
}

func TestAuth_PublicRoute_Health_BypassesAuth(t *testing.T) {
	called := false
	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for /health")
	}
}

func TestAuth_ProtectedRoute_NoToken_Returns401(t *testing.T) {
	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called without token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}

	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Message != "missing or malformed authorization header" {
		t.Errorf("unexpected message: %q", body.Message)
	}
}

func TestAuth_ProtectedRoute_MalformedToken_Returns401(t *testing.T) {
	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with malformed token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-garbage-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ProtectedRoute_ValidToken_SetsHeaders(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	role := "Premium"
	token := generateTestJWT(testSecret, testIssuer, userID, role, time.Now().Add(time.Hour))

	var capturedUserID, capturedRole string
	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = r.Header.Get("X-User-Id")
		capturedRole = r.Header.Get("X-User-Role")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedUserID != userID {
		t.Errorf("expected X-User-Id %q, got %q", userID, capturedUserID)
	}
	if capturedRole != role {
		t.Errorf("expected X-User-Role %q, got %q", role, capturedRole)
	}
}

func TestAuth_ProtectedRoute_ExpiredToken_Returns401(t *testing.T) {
	token := generateTestJWT(testSecret, testIssuer, "user-1", "Free", time.Now().Add(-time.Hour))

	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with expired token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ProtectedRoute_WrongIssuer_Returns401(t *testing.T) {
	token := generateTestJWT(testSecret, "WrongIssuer", "user-1", "Free", time.Now().Add(time.Hour))

	handler := Auth(testSecret, testIssuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with wrong issuer")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestIsPublicRoute(t *testing.T) {
	tests := []struct {
		path   string
		public bool
	}{
		{"/api/auth/login", true},
		{"/api/auth/register", true},
		{"/api/auth/refresh", true},
		{"/api/catalog/movies", true},
		{"/api/catalog/genres", true},
		{"/health", true},
		{"/api/users/me", false},
		{"/api/profiles", false},
		{"/api/watchlist", false},
		{"/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isPublicRoute(tt.path)
			if got != tt.public {
				t.Errorf("isPublicRoute(%q) = %v, want %v", tt.path, got, tt.public)
			}
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantOk    bool
	}{
		{"valid bearer", "Bearer my-token", "my-token", true},
		{"bearer lowercase", "bearer my-token", "my-token", true},
		{"BEARER uppercase", "BEARER my-token", "my-token", true},
		{"empty header", "", "", false},
		{"no bearer prefix", "Token my-token", "", false},
		{"only bearer word", "Bearer", "", false},
		{"bearer with spaces in token", "Bearer token with spaces", "token with spaces", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			gotToken, gotOk := extractBearerToken(req)
			if gotOk != tt.wantOk {
				t.Errorf("extractBearerToken() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && gotToken != tt.wantToken {
				t.Errorf("extractBearerToken() token = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}
