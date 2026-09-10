package metrics

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RED metrics
var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	HTTPRequestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_errors_total",
			Help: "Total number of HTTP requests resulting in errors (4xx/5xx).",
		},
		[]string{"method", "path", "status"},
	)
	HTTPRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)
)

// Business metrics
var (
	NotificationSentTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_sent_total",
			Help: "Total number of notifications sent successfully.",
		},
		[]string{"channel", "category"},
	)
	NotificationFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_failed_total",
			Help: "Total number of notifications that failed to send.",
		},
		[]string{"channel", "reason"},
	)
	NotificationWSActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_ws_active_connections",
			Help: "Number of currently active WebSocket connections.",
		},
	)
	NotificationWSMessagesSentTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_ws_messages_sent_total",
			Help: "Total number of WebSocket messages sent to clients.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestErrorsTotal,
		HTTPRequestDurationSeconds,
		NotificationSentTotal,
		NotificationFailedTotal,
		NotificationWSActiveConnections,
		NotificationWSMessagesSentTotal,
	)
}

// responseRecorder captures status code for metrics. Kept independent from
// the middleware package to avoid import cycles.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Unwrap() http.ResponseWriter {
	return rr.ResponseWriter
}

// Hijack preserves WebSocket upgrade support through the metrics wrapper.
func (rr *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rr.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying response writer does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}

// Middleware returns HTTP middleware that records RED metrics for every request.
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rec.statusCode)
			path := normalizePath(r.URL.Path)

			HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			HTTPRequestDurationSeconds.WithLabelValues(r.Method, path).Observe(duration)

			if rec.statusCode >= 400 {
				HTTPRequestErrorsTotal.WithLabelValues(r.Method, path, status).Inc()
			}
		})
	}
}

// normalizePath collapses dynamic path segments to prevent label cardinality explosion.
func normalizePath(path string) string {
	switch {
	case path == "/health" || path == "/health/live" || path == "/health/ready":
		return "/health"
	case path == "/metrics":
		return "/metrics"
	case strings.HasPrefix(path, "/ws/"):
		return "/ws/notifications"
	case strings.HasPrefix(path, "/api/notifications"):
		return "/api/notifications"
	default:
		return path
	}
}
