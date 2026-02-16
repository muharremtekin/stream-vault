package handler

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	pool         *pgxpool.Pool
	redisClient  *redis.Client
	rabbitCheck  func() bool
	startupReady *atomic.Bool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *pgxpool.Pool, redisClient *redis.Client, rabbitCheck func() bool, startupReady *atomic.Bool) *HealthHandler {
	return &HealthHandler{
		pool:         pool,
		redisClient:  redisClient,
		rabbitCheck:  rabbitCheck,
		startupReady: startupReady,
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

	if h.pool != nil {
		if err := h.pool.Ping(ctx); err != nil {
			log.Warn().Err(err).Msg("postgres health check failed")
			components["postgres"] = componentStatus{Status: "unhealthy"}
			allHealthy = false
		} else {
			components["postgres"] = componentStatus{Status: "healthy"}
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

// ServeStartup handles GET /health/startup (startup probe).
func (h *HealthHandler) ServeStartup(w http.ResponseWriter, r *http.Request) {
	if h.startupReady != nil && h.startupReady.Load() {
		WriteJSON(w, http.StatusOK, healthResponse{Status: "healthy"})
	} else {
		WriteJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "starting"})
	}
}
