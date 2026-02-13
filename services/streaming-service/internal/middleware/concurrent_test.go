package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr, client
}

func TestConcurrentStreams_AllowsFirstRequest(t *testing.T) {
	_, client := setupRedis(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := ConcurrentStreams(client, 5*time.Minute)(handler)

	r := httptest.NewRequest("GET", "/stream/movie-1/720p/seg0.ts", nil)
	r.Header.Set("X-User-Id", "user-1")
	r.Header.Set("X-User-Role", "Premium")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestConcurrentStreams_BlocksWhenLimitExceeded(t *testing.T) {
	mr, client := setupRedis(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := ConcurrentStreams(client, 5*time.Minute)(handler)

	// Free/Basic tier = 1 max stream. Pre-populate with an existing session.
	key := "concurrent:user-free"
	mr.ZAdd(key, float64(time.Now().Unix()), "existing-session")

	r := httptest.NewRequest("GET", "/stream/movie-1/720p/seg0.ts", nil)
	r.Header.Set("X-User-Id", "user-free")
	r.Header.Set("X-User-Role", "Basic")
	r.Header.Set("X-Session-Id", "new-session")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestConcurrentStreams_CleansUpPhantomOnNotFound(t *testing.T) {
	mr, client := setupRedis(t)

	// Handler returns 404 (content not found)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	mw := ConcurrentStreams(client, 5*time.Minute)(handler)

	r := httptest.NewRequest("GET", "/stream/nonexistent/720p/seg0.ts", nil)
	r.Header.Set("X-User-Id", "user-1")
	r.Header.Set("X-User-Role", "Premium")
	r.Header.Set("X-Session-Id", "session-phantom")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	// Verify phantom session was removed from Redis
	// Key may not exist if ZRem removed the only member (miniredis deletes empty sets)
	members, _ := mr.ZMembers("concurrent:user-1")
	for _, m := range members {
		if m == "session-phantom" {
			t.Error("phantom session should have been removed after 404")
		}
	}
}

func TestConcurrentStreams_CleansUpPhantomOnServerError(t *testing.T) {
	mr, client := setupRedis(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	mw := ConcurrentStreams(client, 5*time.Minute)(handler)

	r := httptest.NewRequest("GET", "/stream/movie-1/720p/seg0.ts", nil)
	r.Header.Set("X-User-Id", "user-1")
	r.Header.Set("X-User-Role", "Premium")
	r.Header.Set("X-Session-Id", "session-err")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	// Key may not exist if ZRem removed the only member
	if mr.Exists("concurrent:user-1") {
		t.Error("concurrent key should not exist after cleanup of sole session")
	}
}

func TestConcurrentStreams_KeepsSessionOnSuccess(t *testing.T) {
	mr, client := setupRedis(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("segment-data"))
	})

	mw := ConcurrentStreams(client, 5*time.Minute)(handler)

	r := httptest.NewRequest("GET", "/stream/movie-1/720p/seg0.ts", nil)
	r.Header.Set("X-User-Id", "user-1")
	r.Header.Set("X-User-Role", "Premium")
	r.Header.Set("X-Session-Id", "session-ok")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	members, err := mr.ZMembers("concurrent:user-1")
	if err != nil {
		t.Fatalf("failed to get sorted set members: %v", err)
	}
	found := false
	for _, m := range members {
		if m == "session-ok" {
			found = true
			break
		}
	}
	if !found {
		t.Error("session should remain in concurrent set after successful response")
	}
}

func TestConcurrentStreams_MissingUserID_Returns401(t *testing.T) {
	_, client := setupRedis(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := ConcurrentStreams(client, 5*time.Minute)(handler)

	r := httptest.NewRequest("GET", "/stream/movie-1/720p/seg0.ts", nil)
	// No X-User-Id header
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestConcurrentStreams_ExistingSessionRenewed(t *testing.T) {
	mr, client := setupRedis(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := ConcurrentStreams(client, 5*time.Minute)(handler)

	// Pre-populate with the same session ID (simulates repeated requests from same stream)
	key := "concurrent:user-1"
	mr.ZAdd(key, float64(time.Now().Add(-2*time.Minute).Unix()), "session-renew")

	r := httptest.NewRequest("GET", "/stream/movie-1/720p/seg1.ts", nil)
	r.Header.Set("X-User-Id", "user-1")
	r.Header.Set("X-User-Role", "Basic") // max 1 stream
	r.Header.Set("X-Session-Id", "session-renew")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, r)

	// Should succeed (same session, not a new one)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for renewed session, got %d", w.Code)
	}
}

func TestTierToMaxStreams(t *testing.T) {
	tests := []struct {
		role string
		want int
	}{
		{"Premium", 4},
		{"premium", 4},
		{"Admin", 4},
		{"admin", 4},
		{"Standard", 2},
		{"standard", 2},
		{"Basic", 1},
		{"Free", 1},
		{"", 1},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			got := tierToMaxStreams(tt.role)
			if got != tt.want {
				t.Errorf("tierToMaxStreams(%q) = %d, want %d", tt.role, got, tt.want)
			}
		})
	}
}
