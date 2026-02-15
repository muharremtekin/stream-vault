package metrics

import (
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
	GatewayActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gateway_active_connections",
			Help: "Number of currently active HTTP connections.",
		},
	)
	GatewayRateLimitHitsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_hits_total",
			Help: "Total number of requests rejected by rate limiting.",
		},
	)
	GatewayCircuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_circuit_breaker_state",
			Help: "Circuit breaker state per service (0=closed, 1=half-open, 2=open).",
		},
		[]string{"service", "state"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestErrorsTotal,
		HTTPRequestDurationSeconds,
		GatewayActiveConnections,
		GatewayRateLimitHitsTotal,
		GatewayCircuitBreakerState,
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

// Middleware returns HTTP middleware that records RED metrics for every request.
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			GatewayActiveConnections.Inc()

			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			GatewayActiveConnections.Dec()

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
	case strings.HasPrefix(path, "/api/auth/"):
		return "/api/auth"
	case strings.HasPrefix(path, "/api/catalog/"):
		return "/api/catalog"
	case strings.HasPrefix(path, "/api/stream/"):
		return "/api/stream"
	case strings.HasPrefix(path, "/stream/"):
		return "/stream"
	case strings.HasPrefix(path, "/api/search"):
		return "/api/search"
	case strings.HasPrefix(path, "/api/recommendations"):
		return "/api/recommendations"
	case strings.HasPrefix(path, "/api/notifications"):
		return "/api/notifications"
	case strings.HasPrefix(path, "/api/plans"):
		return "/api/plans"
	case strings.HasPrefix(path, "/api/subscriptions"):
		return "/api/subscriptions"
	case strings.HasPrefix(path, "/api/encoding"):
		return "/api/encoding"
	default:
		return path
	}
}
