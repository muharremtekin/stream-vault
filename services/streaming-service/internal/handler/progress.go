package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/streamvault/streaming-service/internal/progress"
)

type ProgressHandler struct {
	service *progress.Service
}

func NewProgressHandler(svc *progress.Service) *ProgressHandler {
	return &ProgressHandler{service: svc}
}

type saveProgressRequest struct {
	PositionSeconds int64 `json:"position_seconds"`
	DurationSeconds int64 `json:"duration_seconds"`
}

func (h *ProgressHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	contentID := r.PathValue("contentId")
	if contentID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "content_id is required")
		return
	}

	var req saveProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DurationSeconds <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, "duration_seconds must be positive")
		return
	}

	if err := h.service.SaveProgress(r.Context(), userID, contentID, req.PositionSeconds, req.DurationSeconds); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to save progress")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *ProgressHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	contentID := r.PathValue("contentId")
	if contentID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "content_id is required")
		return
	}

	p, err := h.service.GetProgress(r.Context(), userID, contentID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to get progress")
		return
	}

	if p == nil {
		WriteErrorResponse(w, http.StatusNotFound, "no progress found")
		return
	}

	WriteJSON(w, http.StatusOK, p)
}

func (h *ProgressHandler) ContinueWatching(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetContinueWatching(r.Context(), userID, limit)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to get continue watching list")
		return
	}

	if items == nil {
		items = []progress.WatchProgress{}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
	})
}
