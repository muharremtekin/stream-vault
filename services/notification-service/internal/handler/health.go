package handler

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type HealthHandler struct {
	mongoClient    *mongo.Client
	rabbitCheck    func() bool
	hubClientCount func() int
	startupReady   *atomic.Bool
	startTime      time.Time
	serviceName    string
	version        string
}

func NewHealthHandler(mongoClient *mongo.Client, rabbitCheck func() bool, hubClientCount func() int, startupReady *atomic.Bool) *HealthHandler {
	return &HealthHandler{
		mongoClient:    mongoClient,
		rabbitCheck:    rabbitCheck,
		hubClientCount: hubClientCount,
		startupReady:   startupReady,
		startTime:      time.Now(),
		serviceName:    "notification-service",
		version:        "1.0.0",
	}
}

type componentStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Info    string `json:"info,omitempty"`
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

func (h *HealthHandler) ServeReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := make(map[string]componentStatus)
	allHealthy := true

	if h.mongoClient != nil {
		err := h.mongoClient.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
		if err != nil {
			log.Warn().Err(err).Msg("mongodb health check failed")
			components["mongodb"] = componentStatus{Status: "unhealthy"}
			allHealthy = false
		} else {
			components["mongodb"] = componentStatus{Status: "healthy"}
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

	// Check MongoDB
	if h.mongoClient != nil {
		totalCount++
		start := time.Now()
		err := h.mongoClient.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
		if err != nil {
			components["mongodb"] = componentStatus{Status: "unhealthy", Latency: time.Since(start).String()}
			unhealthyCount++
		} else {
			components["mongodb"] = componentStatus{Status: "healthy", Latency: time.Since(start).String()}
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

	// WebSocket hub info (informational, not a health dependency)
	if h.hubClientCount != nil {
		components["websocket"] = componentStatus{
			Status: "healthy",
			Info:   fmt.Sprintf("%d active connections", h.hubClientCount()),
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
