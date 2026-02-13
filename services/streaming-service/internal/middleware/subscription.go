package middleware

import (
	"net/http"

	"github.com/streamvault/streaming-service/internal/handler"
)

func Subscription() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Header.Get("X-User-Role")
			if role == "" {
				role = "Free"
			}

			// Admin has full streaming access
			if role == "Admin" {
				next.ServeHTTP(w, r)
				return
			}

			if role == "Free" || role == "SUBSCRIPTION_TIER_FREE" {
				handler.WriteErrorResponse(w, http.StatusForbidden, "subscription required for streaming")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
