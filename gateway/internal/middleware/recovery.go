package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// Recovery returns middleware that catches panics originating from downstream
// handlers, logs the stack trace, and responds with 500 Internal Server Error
// instead of crashing the entire process.
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()

					log.Error().
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("panic", fmt.Sprintf("%v", rec)).
						Str("stack", string(stack)).
						Msg("recovered from panic")

					WriteErrorResponse(w, http.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
