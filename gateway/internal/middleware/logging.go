package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

var sensitiveQueryParameters = map[string]struct{}{
	"access_token":  {},
	"api_key":       {},
	"password":      {},
	"refresh_token": {},
	"secret":        {},
	"token":         {},
}

func redactSensitiveQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "[invalid query]"
	}

	for key := range values {
		if _, sensitive := sensitiveQueryParameters[strings.ToLower(key)]; sensitive {
			values.Set(key, "[REDACTED]")
		}
	}

	return values.Encode()
}

// responseRecorder wraps http.ResponseWriter to capture the status code and
// bytes written for logging.
type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// newResponseRecorder creates a recorder that defaults to 200 OK.
func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code before delegating to the wrapped writer.
func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// Write captures the number of bytes written.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	n, err := rr.ResponseWriter.Write(b)
	rr.bytesWritten += n
	return n, err
}

// Unwrap supports http.ResponseController by exposing the underlying writer.
func (rr *responseRecorder) Unwrap() http.ResponseWriter {
	return rr.ResponseWriter
}

// Logging returns middleware that emits a structured JSON log line for every
// HTTP request/response pair using zerolog. It records method, path, status,
// latency, client IP, and response size.
func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := newResponseRecorder(w)

			next.ServeHTTP(rec, r)

			duration := time.Since(start)

			event := log.Info()
			if rec.statusCode >= 500 {
				event = log.Error()
			} else if rec.statusCode >= 400 {
				event = log.Warn()
			}

			// Extract trace_id and span_id from OTel span context.
			traceID := ""
			spanID := ""
			span := trace.SpanFromContext(r.Context())
			if span.SpanContext().HasTraceID() {
				traceID = span.SpanContext().TraceID().String()
			}
			if span.SpanContext().HasSpanID() {
				spanID = span.SpanContext().SpanID().String()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("query", redactSensitiveQuery(r.URL.RawQuery)).
				Int("status", rec.statusCode).
				Dur("latency", duration).
				Int("response_bytes", rec.bytesWritten).
				Str("client_ip", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Str("correlation_id", r.Header.Get("X-Correlation-Id")).
				Str("trace_id", traceID).
				Str("span_id", spanID).
				Msg("request completed")
		})
	}
}
