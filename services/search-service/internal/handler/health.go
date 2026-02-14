package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/elasticsearch"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	esClient    *elasticsearch.Client
	redisClient *redis.Client
	rabbitCheck func() bool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(esClient *elasticsearch.Client, redisClient *redis.Client, rabbitCheck func() bool) *HealthHandler {
	return &HealthHandler{
		esClient:    esClient,
		redisClient: redisClient,
		rabbitCheck: rabbitCheck,
	}
}

type componentStatus struct {
	Status string `json:"status"`
}

type healthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]componentStatus `json:"components,omitempty"`
}

// ServeLive handles GET /health and GET /health/live (liveness probe).
func (h *HealthHandler) ServeLive(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, healthResponse{Status: "healthy"})
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
