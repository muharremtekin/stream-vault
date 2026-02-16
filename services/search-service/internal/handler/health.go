package handler

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/elasticsearch"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	esClient     *elasticsearch.Client
	redisClient  *redis.Client
	rabbitCheck  func() bool
	startupReady *atomic.Bool
	startTime    time.Time
	serviceName  string
	version      string
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(esClient *elasticsearch.Client, redisClient *redis.Client, rabbitCheck func() bool, startupReady *atomic.Bool) *HealthHandler {
	return &HealthHandler{
		esClient:     esClient,
		redisClient:  redisClient,
		rabbitCheck:  rabbitCheck,
		startupReady: startupReady,
		startTime:    time.Now(),
		serviceName:  "search-service",
		version:      "1.0.0",
	}
}

type componentStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
}

type healthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]componentStatus `json:"components,omitempty"`
}

type detailedHealthResponse struct {
	Status           string                     `json:"status"`
	Service          string                     `json:"service"`
	Version          string                     `json:"version"`
	Uptime           string                     `json:"uptime"`
	StartupCompleted bool                       `json:"startupCompleted"`
	Components       map[string]componentStatus `json:"components"`
}

// ServeLive handles GET /health and GET /health/live (liveness probe).
func (h *HealthHandler) ServeLive(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, healthResponse{Status: "healthy"})
}

// ServeStartup handles GET /health/startup requests (startup probe).
// Returns 200 once all initialization is complete, 503 while starting.
func (h *HealthHandler) ServeStartup(w http.ResponseWriter, r *http.Request) {
	if h.startupReady != nil && h.startupReady.Load() {
		WriteJSON(w, http.StatusOK, healthResponse{Status: "healthy"})
		return
	}
	WriteJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "starting"})
}

// ServeReady handles GET /health/ready (readiness probe).
func (h *HealthHandler) ServeReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := make(map[string]componentStatus)
	allHealthy := true

	if h.esClient != nil {
		if err := h.esClient.HealthCheck(ctx); err != nil {
			log.Warn().Err(err).Msg("elasticsearch health check failed")
			components["elasticsearch"] = componentStatus{Status: "unhealthy"}
			allHealthy = false
		} else {
			components["elasticsearch"] = componentStatus{Status: "healthy"}
		}
	}

	if h.redisClient != nil {
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			log.Warn().Err(err).Msg("redis health check failed")
			components["redis"] = componentStatus{Status: "unhealthy"}
			allHealthy = false
		} else {
			components["redis"] = componentStatus{Status: "healthy"}
		}
	}

	if h.rabbitCheck != nil {
		if h.rabbitCheck() {
			components["rabbitmq"] = componentStatus{Status: "healthy"}
		} else {
			components["rabbitmq"] = componentStatus{Status: "unhealthy"}
			allHealthy = false
		}
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	WriteJSON(w, httpStatus, healthResponse{
		Status:     status,
		Components: components,
	})
}

// ServeDetailed handles GET /health requests with a full diagnostic report.
// Includes version, uptime, startup status, and all dependency checks with latency.
func (h *HealthHandler) ServeDetailed(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := make(map[string]componentStatus)
	unhealthyCount := 0
	totalCount := 0

	// Check Elasticsearch
	if h.esClient != nil {
		totalCount++
		start := time.Now()
		if err := h.esClient.HealthCheck(ctx); err != nil {
			components["elasticsearch"] = componentStatus{Status: "unhealthy", Latency: time.Since(start).String()}
			unhealthyCount++
		} else {
			components["elasticsearch"] = componentStatus{Status: "healthy", Latency: time.Since(start).String()}
		}
	}

	// Check Redis
	if h.redisClient != nil {
		totalCount++
		start := time.Now()
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			components["redis"] = componentStatus{Status: "unhealthy", Latency: time.Since(start).String()}
			unhealthyCount++
		} else {
			components["redis"] = componentStatus{Status: "healthy", Latency: time.Since(start).String()}
		}
	}

	// Check RabbitMQ
	if h.rabbitCheck != nil {
		totalCount++
		if h.rabbitCheck() {
			components["rabbitmq"] = componentStatus{Status: "healthy"}
		} else {
			components["rabbitmq"] = componentStatus{Status: "unhealthy"}
			unhealthyCount++
		}
	}

	status := "healthy"
	if unhealthyCount > 0 && unhealthyCount < totalCount {
		status = "degraded"
	} else if unhealthyCount == totalCount && totalCount > 0 {
		status = "unhealthy"
	}

	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	WriteJSON(w, httpStatus, detailedHealthResponse{
		Status:           status,
		Service:          h.serviceName,
		Version:          h.version,
		Uptime:           time.Since(h.startTime).Round(time.Second).String(),
		StartupCompleted: h.startupReady != nil && h.startupReady.Load(),
		Components:       components,
	})
}
