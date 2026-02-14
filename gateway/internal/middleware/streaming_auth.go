package middleware

import (
	"net/http"
	"strings"
)

// adminPrefixes lists URL path prefixes that require the Admin role.
var adminPrefixes = []string{
	"/api/stream/upload",
	"/api/encoding/",
}

// StreamingAuth returns middleware that enforces Admin role checks on
// privileged routes such as video upload and encoding job management.
// Routes that do not match any admin prefix pass through unchanged.
func StreamingAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requiresAdmin(r.URL.Path) {
				role := r.Header.Get("X-User-Role")
				if role != "Admin" {
					WriteErrorResponse(w, http.StatusForbidden, "admin access required")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// requiresAdmin checks whether the given path matches any admin-only prefix.
func requiresAdmin(path string) bool {
	for _, prefix := range adminPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
