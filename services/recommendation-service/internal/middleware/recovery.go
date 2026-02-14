package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/handler"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error().
						Interface("panic", rec).
						Str("stack", string(debug.Stack())).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("panic recovered")
					handler.WriteErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
