package middleware

import (
	"net/http"

	"github.com/streamvault/streaming-service/internal/handler"
)

func Admin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Header.Get("X-User-Role")
			if role != "Admin" {
				handler.WriteErrorResponse(w, http.StatusForbidden, "admin access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
