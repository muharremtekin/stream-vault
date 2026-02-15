package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/notification-service/internal/store"
)

// NotificationHandler handles notification history and preference endpoints.
type NotificationHandler struct {
	notifStore store.NotificationStore
	prefStore  store.PreferencesStore
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(notifStore store.NotificationStore, prefStore store.PreferencesStore) *NotificationHandler {
	return &NotificationHandler{
		notifStore: notifStore,
		prefStore:  prefStore,
	}
}

// notificationListResponse is the paginated response for listing notifications.
type notificationListResponse struct {
	Items       []store.Notification `json:"items"`
	UnreadCount int64                `json:"unreadCount"`
	TotalCount  int64                `json:"totalCount"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"pageSize"`
	TotalPages  int                  `json:"totalPages"`
}

// ServeListNotifications handles GET /api/notifications.
func (h *NotificationHandler) ServeListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_USER_ID", "X-User-Id header is required")
		return
	}

	page, pageSize := parsePagination(r)
	unreadOnly := parseUnreadOnly(r)

	notifications, totalCount, err := h.notifStore.ListByUserID(r.Context(), userID, page, pageSize, unreadOnly)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to list notifications")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list notifications")
		return
	}

	unreadCount, err := h.notifStore.CountUnread(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to count unread notifications")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to count unread notifications")
		return
	}

	if notifications == nil {
		notifications = []store.Notification{}
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	WriteJSON(w, http.StatusOK, notificationListResponse{
		Items:       notifications,
		UnreadCount: unreadCount,
		TotalCount:  totalCount,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	})
}

// ServeMarkAsRead handles POST /api/notifications/{id}/read.
func (h *NotificationHandler) ServeMarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_USER_ID", "X-User-Id header is required")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_ID", "notification id path parameter is required")
		return
	}

	err := h.notifStore.MarkAsRead(r.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteErrorResponse(w, r, http.StatusNotFound, "NOT_FOUND", "notification not found")
			return
		}
		if strings.Contains(err.Error(), "invalid notification id") {
			WriteErrorResponse(w, r, http.StatusBadRequest, "INVALID_ID", "invalid notification id format")
			return
		}
		log.Error().Err(err).Str("notification_id", id).Str("user_id", userID).Msg("failed to mark notification as read")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to mark notification as read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ServeMarkAllAsRead handles POST /api/notifications/read-all.
func (h *NotificationHandler) ServeMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_USER_ID", "X-User-Id header is required")
		return
	}

	err := h.notifStore.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to mark all notifications as read")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to mark all notifications as read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ServeGetPreferences handles GET /api/notifications/preferences.
func (h *NotificationHandler) ServeGetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_USER_ID", "X-User-Id header is required")
		return
	}

	prefs, err := h.prefStore.Get(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to get notification preferences")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get notification preferences")
		return
	}

	if prefs == nil {
		prefs = store.DefaultPreferences(userID)
	}

	WriteJSON(w, http.StatusOK, prefs)
}

// updatePreferencesRequest is the request body for updating notification preferences.
type updatePreferencesRequest struct {
	Email *store.ChannelPreference `json:"email"`
	Push  *store.ChannelPreference `json:"push"`
	InApp *store.ChannelPreference `json:"inApp"`
}

// ServeUpdatePreferences handles PUT /api/notifications/preferences.
func (h *NotificationHandler) ServeUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_USER_ID", "X-User-Id header is required")
		return
	}

	var req updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, r, http.StatusBadRequest, "INVALID_BODY", "invalid JSON request body")
		return
	}

	if req.Email == nil && req.Push == nil && req.InApp == nil {
		WriteValidationError(w, r, []FieldError{
			{Field: "body", Message: "at least one preference field (email, push, inApp) must be provided"},
		})
		return
	}

	// Fetch existing preferences to merge with the update.
	existing, err := h.prefStore.Get(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to get existing preferences")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update preferences")
		return
	}
	if existing == nil {
		existing = store.DefaultPreferences(userID)
	}

	// Overlay non-nil fields from the request.
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Push != nil {
		existing.Push = *req.Push
	}
	if req.InApp != nil {
		existing.InApp = *req.InApp
	}

	if err := h.prefStore.Upsert(r.Context(), existing); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to upsert preferences")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update preferences")
		return
	}

	WriteJSON(w, http.StatusOK, existing)
}

// parsePagination extracts page and pageSize from query parameters with defaults.
func parsePagination(r *http.Request) (int, int) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	return page, pageSize
}

// parseUnreadOnly extracts the unreadOnly boolean from query parameters.
func parseUnreadOnly(r *http.Request) bool {
	val := r.URL.Query().Get("unreadOnly")
	return val == "true" || val == "1"
}
