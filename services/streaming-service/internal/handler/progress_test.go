package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/streamvault/streaming-service/internal/progress"
)

func newProgressHandler(repo progress.Repository) *ProgressHandler {
	svc := progress.NewService(repo)
	return NewProgressHandler(svc)
}

func TestSaveProgress_NegativePosition_Returns400(t *testing.T) {
	h := newProgressHandler(&mockProgressRepo{})

	body := `{"position_seconds": -10, "duration_seconds": 100}`
	req := httptest.NewRequest(http.MethodPost, "/stream/{contentId}/progress", strings.NewReader(body))
	req.Header.Set("X-User-Id", "user-1")
	req.SetPathValue("contentId", "movie-1")

	rr := httptest.NewRecorder()
	h.SaveProgress(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for negative positionSeconds, got %d", rr.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if !strings.Contains(resp.Message, "position_seconds") {
		t.Errorf("expected error message about position_seconds, got: %s", resp.Message)
	}
}

func TestSaveProgress_ZeroPosition_Returns200(t *testing.T) {
	h := newProgressHandler(&mockProgressRepo{})

	body := `{"position_seconds": 0, "duration_seconds": 100}`
	req := httptest.NewRequest(http.MethodPost, "/stream/{contentId}/progress", strings.NewReader(body))
	req.Header.Set("X-User-Id", "user-1")
	req.SetPathValue("contentId", "movie-1")

	rr := httptest.NewRecorder()
	h.SaveProgress(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for zero positionSeconds, got %d", rr.Code)
	}
}

func TestSaveProgress_ValidPosition_Returns200(t *testing.T) {
	saved := false
	h := newProgressHandler(&mockProgressRepo{
		saveFunc: func(_ context.Context, p progress.WatchProgress) error {
			saved = true
			if p.PositionSeconds != 50 {
				t.Errorf("expected positionSeconds=50, got %d", p.PositionSeconds)
			}
			return nil
		},
	})

	body := `{"position_seconds": 50, "duration_seconds": 100}`
	req := httptest.NewRequest(http.MethodPost, "/stream/{contentId}/progress", strings.NewReader(body))
	req.Header.Set("X-User-Id", "user-1")
	req.SetPathValue("contentId", "movie-1")

	rr := httptest.NewRecorder()
	h.SaveProgress(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if !saved {
		t.Error("expected Save to be called for valid progress")
	}
}

func TestSaveProgress_ZeroDuration_Returns400(t *testing.T) {
	h := newProgressHandler(&mockProgressRepo{})

	body := `{"position_seconds": 10, "duration_seconds": 0}`
	req := httptest.NewRequest(http.MethodPost, "/stream/{contentId}/progress", strings.NewReader(body))
	req.Header.Set("X-User-Id", "user-1")
	req.SetPathValue("contentId", "movie-1")

	rr := httptest.NewRecorder()
	h.SaveProgress(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for zero duration, got %d", rr.Code)
	}
}

func TestSaveProgress_MissingUserID_Returns401(t *testing.T) {
	h := newProgressHandler(&mockProgressRepo{})

	body := `{"position_seconds": 10, "duration_seconds": 100}`
	req := httptest.NewRequest(http.MethodPost, "/stream/{contentId}/progress", strings.NewReader(body))
	req.SetPathValue("contentId", "movie-1")

	rr := httptest.NewRecorder()
	h.SaveProgress(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}
