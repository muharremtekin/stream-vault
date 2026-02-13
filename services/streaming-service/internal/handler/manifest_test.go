package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/storage"
)

func TestMediaPlaylist_Success_StreamsContent(t *testing.T) {
	playlistData := "#EXTM3U\n#EXT-X-VERSION:3\n#EXTINF:10,\nseg0.ts\n"
	store := &mockStorage{
		downloadFunc: func(_ context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
			return io.NopCloser(strings.NewReader(playlistData)), &storage.ObjectInfo{
				Key:  key,
				Size: int64(len(playlistData)),
			}, nil
		},
	}

	h := NewManifestHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/720p/playlist.m3u8", nil)
	req.SetPathValue("contentId", "movie-1")
	req.SetPathValue("quality", "720p")

	rr := httptest.NewRecorder()
	h.MediaPlaylist(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/vnd.apple.mpegurl" {
		t.Errorf("expected Content-Type application/vnd.apple.mpegurl, got %s", ct)
	}
	if body := rr.Body.String(); body != playlistData {
		t.Errorf("expected body %q, got %q", playlistData, body)
	}
}

func TestMediaPlaylist_MissingParams_Returns400(t *testing.T) {
	h := NewManifestHandler(&mockStorage{}, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream//", nil)
	req.SetPathValue("contentId", "")
	req.SetPathValue("quality", "")

	rr := httptest.NewRecorder()
	h.MediaPlaylist(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestMasterPlaylist_Success_ContainsAbsoluteURIs(t *testing.T) {
	store := &mockStorage{
		listFunc: func(_ context.Context, bucket, prefix string) ([]storage.ObjectInfo, error) {
			return []storage.ObjectInfo{
				{Key: "movie-1/720p/playlist.m3u8"},
				{Key: "movie-1/720p/seg0.ts"},
				{Key: "movie-1/1080p/playlist.m3u8"},
			}, nil
		},
	}

	h := NewManifestHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.SetPathValue("contentId", "movie-1")
	req.Header.Set("X-User-Role", "Premium")

	rr := httptest.NewRecorder()
	h.MasterPlaylist(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rr.Code, rr.Body.String())
	}

	body := rr.Body.String()
	if !strings.Contains(body, "/stream/movie-1/720p/playlist.m3u8") {
		t.Errorf("expected absolute URI with contentId for 720p in body:\n%s", body)
	}
}

func TestMediaPlaylist_NotFound_Returns404(t *testing.T) {
	store := &mockStorage{
		downloadFunc: func(_ context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
			return nil, nil, fmt.Errorf("stat object: %w", minio.ErrorResponse{Code: "NoSuchKey"})
		},
	}

	h := NewManifestHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/720p/playlist.m3u8", nil)
	req.SetPathValue("contentId", "movie-1")
	req.SetPathValue("quality", "720p")

	rr := httptest.NewRecorder()
	h.MediaPlaylist(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for NoSuchKey, got %d", rr.Code)
	}
}

func TestMediaPlaylist_StorageError_Returns502(t *testing.T) {
	store := &mockStorage{
		downloadFunc: func(_ context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
			return nil, nil, fmt.Errorf("getting object: %w", errors.New("connection timeout"))
		},
	}

	h := NewManifestHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/720p/playlist.m3u8", nil)
	req.SetPathValue("contentId", "movie-1")
	req.SetPathValue("quality", "720p")

	rr := httptest.NewRecorder()
	h.MediaPlaylist(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected status 502 for storage error, got %d", rr.Code)
	}
}

func TestMasterPlaylist_NoContent_Returns404(t *testing.T) {
	store := &mockStorage{
		listFunc: func(_ context.Context, bucket, prefix string) ([]storage.ObjectInfo, error) {
			return nil, nil
		},
	}

	h := NewManifestHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/nonexistent/manifest.m3u8", nil)
	req.SetPathValue("contentId", "nonexistent")
	req.Header.Set("X-User-Role", "Premium")

	rr := httptest.NewRecorder()
	h.MasterPlaylist(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestMasterPlaylist_FreeTier_Returns403(t *testing.T) {
	store := &mockStorage{
		listFunc: func(_ context.Context, bucket, prefix string) ([]storage.ObjectInfo, error) {
			return []storage.ObjectInfo{
				{Key: "movie-1/1080p/seg0.ts"},
				{Key: "movie-1/4k/seg0.ts"},
			}, nil
		},
	}

	h := NewManifestHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/manifest.m3u8", nil)
	req.SetPathValue("contentId", "movie-1")
	req.Header.Set("X-User-Role", "Free")

	rr := httptest.NewRecorder()
	h.MasterPlaylist(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}
