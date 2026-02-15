package metrics

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
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

// gRPC metrics
var (
	GRPCRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"method", "status"},
	)
	GRPCRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method"},
	)
)

// Business metrics
var (
	RecommendationRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recommendation_requests_total",
			Help: "Total recommendation requests by algorithm type.",
		},
		[]string{"algorithm"},
	)
	RecommendationComputationDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "recommendation_computation_duration_seconds",
			Help:    "Recommendation computation duration in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"algorithm"},
	)
	RecommendationCacheHitsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "recommendation_cache_hits_total",
			Help: "Total recommendation cache hits.",
		},
	)
	RecommendationCacheMissesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "recommendation_cache_misses_total",
			Help: "Total recommendation cache misses.",
		},
	)
	RecommendationColdStartFallbackTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "recommendation_cold_start_fallback_total",
			Help: "Total times the cold start fallback to popularity was triggered.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestErrorsTotal,
		HTTPRequestDurationSeconds,
		GRPCRequestsTotal,
		GRPCRequestDurationSeconds,
		RecommendationRequestsTotal,
		RecommendationComputationDurationSeconds,
		RecommendationCacheHitsTotal,
		RecommendationCacheMissesTotal,
		RecommendationColdStartFallbackTotal,
	)
}

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

// Middleware returns HTTP middleware that records RED metrics.
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
			statusStr := strconv.Itoa(rec.statusCode)
			path := normalizePath(r.URL.Path)

			HTTPRequestsTotal.WithLabelValues(r.Method, path, statusStr).Inc()
			HTTPRequestDurationSeconds.WithLabelValues(r.Method, path).Observe(duration)

			if rec.statusCode >= 400 {
				HTTPRequestErrorsTotal.WithLabelValues(r.Method, path, statusStr).Inc()
			}
		})
	}
}

// GRPCUnaryInterceptor returns a gRPC unary server interceptor for metrics.
func GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		st, _ := status.FromError(err)
		GRPCRequestsTotal.WithLabelValues(info.FullMethod, st.Code().String()).Inc()
		GRPCRequestDurationSeconds.WithLabelValues(info.FullMethod).Observe(duration)
		return resp, err
	}
}

func normalizePath(path string) string {
	switch {
	case path == "/health" || path == "/health/live" || path == "/health/ready":
		return "/health"
	case path == "/metrics":
		return "/metrics"
	case strings.HasPrefix(path, "/api/recommendations/similar/"):
		return "/api/recommendations/similar/{contentId}"
	case strings.HasPrefix(path, "/api/recommendations"):
		return "/api/recommendations"
	default:
		return path
	}
}
