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

func TestServeSegment_Success_StreamsContent(t *testing.T) {
	segmentData := "fake-segment-data"
	store := &mockStorage{
		downloadFunc: func(_ context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
			return io.NopCloser(strings.NewReader(segmentData)), &storage.ObjectInfo{
				Key:  key,
				Size: int64(len(segmentData)),
			}, nil
		},
	}

	h := NewSegmentHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/720p/seg0.ts", nil)
	req.SetPathValue("contentId", "movie-1")
	req.SetPathValue("quality", "720p")
	req.SetPathValue("segment", "seg0.ts")

	rr := httptest.NewRecorder()
	h.ServeSegment(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "video/mp2t" {
		t.Errorf("expected Content-Type video/mp2t, got %s", ct)
	}
	if body := rr.Body.String(); body != segmentData {
		t.Errorf("expected body %q, got %q", segmentData, body)
	}
}

func TestServeSegment_MissingParams_Returns400(t *testing.T) {
	h := NewSegmentHandler(&mockStorage{}, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream///", nil)
	req.SetPathValue("contentId", "")
	req.SetPathValue("quality", "")
	req.SetPathValue("segment", "")

	rr := httptest.NewRecorder()
	h.ServeSegment(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestServeSegment_NotFound_Returns404(t *testing.T) {
	store := &mockStorage{
		downloadFunc: func(_ context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
			return nil, nil, fmt.Errorf("stat object %s/%s: %w", bucket, key, minio.ErrorResponse{Code: "NoSuchKey"})
		},
	}

	h := NewSegmentHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/720p/seg99.ts", nil)
	req.SetPathValue("contentId", "movie-1")
	req.SetPathValue("quality", "720p")
	req.SetPathValue("segment", "seg99.ts")

	rr := httptest.NewRecorder()
	h.ServeSegment(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for NoSuchKey, got %d", rr.Code)
	}
}

func TestServeSegment_StorageError_Returns502(t *testing.T) {
	store := &mockStorage{
		downloadFunc: func(_ context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
			return nil, nil, fmt.Errorf("getting object: %w", errors.New("connection refused"))
		},
	}

	h := NewSegmentHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/720p/seg0.ts", nil)
	req.SetPathValue("contentId", "movie-1")
	req.SetPathValue("quality", "720p")
	req.SetPathValue("segment", "seg0.ts")

	rr := httptest.NewRecorder()
	h.ServeSegment(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected status 502 for storage error, got %d", rr.Code)
	}
}

func TestServeSegment_AccessDenied_Returns502(t *testing.T) {
	store := &mockStorage{
		downloadFunc: func(_ context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
			return nil, nil, fmt.Errorf("getting object: %w", minio.ErrorResponse{Code: "AccessDenied"})
		},
	}

	h := NewSegmentHandler(store, config.MinIOConfig{EncodedBucket: "encoded"})

	req := httptest.NewRequest(http.MethodGet, "/stream/movie-1/720p/seg0.ts", nil)
	req.SetPathValue("contentId", "movie-1")
	req.SetPathValue("quality", "720p")
	req.SetPathValue("segment", "seg0.ts")

	rr := httptest.NewRecorder()
	h.ServeSegment(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected status 502 for AccessDenied (not 404), got %d", rr.Code)
	}
}
