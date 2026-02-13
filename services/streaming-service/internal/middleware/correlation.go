package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const CorrelationHeader = "X-Correlation-Id"

func CorrelationID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			correlationID := r.Header.Get(CorrelationHeader)
			if correlationID == "" {
				correlationID = uuid.New().String()
				r.Header.Set(CorrelationHeader, correlationID)
			}
			w.Header().Set(CorrelationHeader, correlationID)
			next.ServeHTTP(w, r)
		})
	}
}
