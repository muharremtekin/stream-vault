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

			// Enrich zerolog context with correlation_id and trace_id.
			logCtx := log.With().Str("correlation_id", correlationID)
			span := trace.SpanFromContext(r.Context())
			if span.SpanContext().HasTraceID() {
				logCtx = logCtx.Str("trace_id", span.SpanContext().TraceID().String())
			}
			logger := logCtx.Logger()
			ctx := logger.WithContext(r.Context())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
