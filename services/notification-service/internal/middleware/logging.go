package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += n
	return n, err
}

func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := newResponseRecorder(w)

			next.ServeHTTP(rec, r)

			latency := time.Since(start)
			traceID := ""
			span := trace.SpanFromContext(r.Context())
			if span.SpanContext().HasTraceID() {
				traceID = span.SpanContext().TraceID().String()
			}

			logger := log.With().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rec.statusCode).
				Dur("latency", latency).
				Int("bytes", rec.bytesWritten).
				Str("correlation_id", r.Header.Get(CorrelationHeader)).
				Str("trace_id", traceID).
				Logger()

			switch {
			case rec.statusCode >= 500:
				logger.Error().Msg("request completed")
			case rec.statusCode >= 400:
				logger.Warn().Msg("request completed")
			default:
				logger.Info().Msg("request completed")
			}
		})
	}
}
