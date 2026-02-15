package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

const correlationIDHeader = "X-Correlation-Id"

// CorrelationID returns middleware that ensures every request has an
// X-Correlation-Id header. If the client does not provide one a new UUID is
// generated. The ID is set on the request (propagated to upstreams), on the
// response (returned to client), and added to the zerolog context.
func CorrelationID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cid := r.Header.Get(correlationIDHeader)
			if cid == "" {
				cid = generateUUID()
			}

			// Set on request so upstream services receive it.
			r.Header.Set(correlationIDHeader, cid)

			// Set on response so the client can see it.
			w.Header().Set(correlationIDHeader, cid)

			// Add to zerolog context for structured logging.
			logCtx := log.With().Str("correlation_id", cid)

			// Enrich log context with trace_id from OpenTelemetry span.
			span := trace.SpanFromContext(r.Context())
			if span.SpanContext().HasTraceID() {
				logCtx = logCtx.Str("trace_id", span.SpanContext().TraceID().String())
			}

			logger := logCtx.Logger()
			ctx := logger.WithContext(r.Context())

			// Also store in context via zerolog so downstream code
			// using zerolog.Ctx(ctx) gets the enriched logger.
			ctx = zerolog.Ctx(ctx).WithContext(ctx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
