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
	StreamingActiveViewers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "streaming_active_viewers",
			Help: "Number of currently active streaming viewers.",
		},
	)
	StreamingConcurrentViewers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "streaming_concurrent_viewers",
			Help: "Number of concurrent viewers by subscription tier.",
		},
		[]string{"tier"},
	)
	StreamingSegmentServeDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "streaming_segment_serve_duration_seconds",
			Help:    "Time to serve a video segment.",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
		},
		[]string{"quality"},
	)
	StreamingProgressSavesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "streaming_progress_saves_total",
			Help: "Total number of watch progress saves.",
		},
	)
	StreamingBandwidthBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "streaming_bandwidth_bytes_total",
			Help: "Total bytes streamed to clients.",
		},
		[]string{"quality"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestErrorsTotal,
		HTTPRequestDurationSeconds,
		GRPCRequestsTotal,
		GRPCRequestDurationSeconds,
		StreamingActiveViewers,
		StreamingConcurrentViewers,
		StreamingSegmentServeDuration,
		StreamingProgressSavesTotal,
		StreamingBandwidthBytesTotal,
	)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	n, err := rr.ResponseWriter.Write(b)
	rr.bytesWritten += n
	return n, err
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
	case strings.HasPrefix(path, "/api/stream/upload"):
		return "/api/stream/upload"
	case strings.HasPrefix(path, "/api/stream/continue-watching"):
		return "/api/stream/continue-watching"
	case strings.HasPrefix(path, "/api/stream/") && strings.HasSuffix(path, "/progress"):
		return "/api/stream/{contentId}/progress"
	case strings.HasPrefix(path, "/stream/"):
		return "/stream/{contentId}/{quality}/{segment}"
	default:
		return path
	}
}
