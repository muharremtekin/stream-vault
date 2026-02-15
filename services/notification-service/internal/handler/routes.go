package handler

import (
	"net/http"

	"github.com/streamvault/notification-service/internal/store"
)

// RegisterNotificationRoutes registers all notification API routes on the mux.
func RegisterNotificationRoutes(
	mux *http.ServeMux,
	notifStore store.NotificationStore,
	prefStore store.PreferencesStore,
) {
	h := NewNotificationHandler(notifStore, prefStore)

	mux.HandleFunc("GET /api/notifications", h.ServeListNotifications)
	mux.HandleFunc("POST /api/notifications/{id}/read", h.ServeMarkAsRead)
	mux.HandleFunc("POST /api/notifications/read-all", h.ServeMarkAllAsRead)
	mux.HandleFunc("GET /api/notifications/preferences", h.ServeGetPreferences)
	mux.HandleFunc("PUT /api/notifications/preferences", h.ServeUpdatePreferences)
}
