package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/storage"
)

type HealthHandler struct {
	storage     storage.Storage
	redisClient *redis.Client
	rabbitCheck func() bool
}

func NewHealthHandler(store storage.Storage, redisClient *redis.Client, rabbitCheck func() bool) *HealthHandler {
	return &HealthHandler{
		storage:     store,
		redisClient: redisClient,
		rabbitCheck: rabbitCheck,
	}
}

type componentStatus struct {
	Status string `json:"status"`
}

type healthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]componentStatus `json:"components"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := make(map[string]componentStatus)
	allHealthy := true

	// Check MinIO
	if h.storage != nil {
		if err := h.storage.HealthCheck(ctx); err != nil {
			log.Warn().Err(err).Msg("minio health check failed")
			components["minio"] = componentStatus{Status: "unhealthy"}
			allHealthy = false
		} else {
			components["minio"] = componentStatus{Status: "healthy"}
		}
	}

	// Check Redis
	if h.redisClient != nil {
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			log.Warn().Err(err).Msg("redis health check failed")
			components["redis"] = componentStatus{Status: "unhealthy"}
			allHealthy = false
		} else {
			components["redis"] = componentStatus{Status: "healthy"}
		}
	}

	// Check RabbitMQ
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
