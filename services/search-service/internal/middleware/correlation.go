package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
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

			// Add trace_id from OTel span context into zerolog logger context
			span := trace.SpanFromContext(r.Context())
			traceID := span.SpanContext().TraceID().String()
			logger := log.With().
				Str("correlation_id", correlationID).
				Str("trace_id", traceID).
				Logger()
			r = r.WithContext(logger.WithContext(r.Context()))

			next.ServeHTTP(w, r)
		})
	}
}
