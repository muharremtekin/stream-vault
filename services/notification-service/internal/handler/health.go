package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type HealthHandler struct {
	mongoClient *mongo.Client
	rabbitCheck func() bool
}

func NewHealthHandler(mongoClient *mongo.Client, rabbitCheck func() bool) *HealthHandler {
	return &HealthHandler{
		mongoClient: mongoClient,
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

func (h *HealthHandler) ServeLive(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, healthResponse{Status: "healthy"})
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
