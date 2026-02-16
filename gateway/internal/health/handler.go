package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streamvault/gateway/internal/config"
	"github.com/streamvault/gateway/internal/discovery"
)

// Status represents the health state of a single service.
type Status struct {
	Service string `json:"service"`
	Status  string `json:"status"` // "healthy" or "unhealthy"
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Response is the aggregated health check response returned by the gateway.
type Response struct {
	Status   string   `json:"status"` // "healthy", "degraded", or "unhealthy"
	Gateway  string   `json:"gateway"`
	Upstream []Status `json:"upstream,omitempty"`
}

// DetailedResponse is the full diagnostic health report returned by GET /health.
type DetailedResponse struct {
	Status           string   `json:"status"`
	Service          string   `json:"service"`
	Version          string   `json:"version"`
	Uptime           string   `json:"uptime"`
	StartupCompleted bool     `json:"startupCompleted"`
	Gateway          string   `json:"gateway"`
	Upstream         []Status `json:"upstream"`
}

// Handler serves the health endpoints. It probes every configured downstream
// service and aggregates the results into a single response.
type Handler struct {
	services     map[string]config.ServiceEntry
	resolver     *discovery.Resolver
	client       *http.Client
	startupReady *atomic.Bool
	startTime    time.Time
	version      string
	redisCheck   func() bool
}

// NewHandler creates a health Handler that checks the provided services.
func NewHandler(services map[string]config.ServiceEntry, resolver *discovery.Resolver, startupReady *atomic.Bool, redisCheck func() bool) *Handler {
	return &Handler{
		services:     services,
		resolver:     resolver,
		client:       &http.Client{Timeout: 5 * time.Second},
		startupReady: startupReady,
		startTime:    time.Now(),
		version:      "1.0.0",
		redisCheck:   redisCheck,
	}
}

// ServeHTTP handles GET /health requests. Returns a detailed health report.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeDetailed(w, r)
}

// ServeLive handles GET /health/live requests (liveness probe).
// Returns 200 if the gateway process is alive — no upstream checks.
func (h *Handler) ServeLive(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Status:  "healthy",
		Gateway: "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("failed to encode health response")
	}
}

// ServeStartup handles GET /health/startup requests (startup probe).
// Returns 200 once all initialization is complete, 503 while starting.
func (h *Handler) ServeStartup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.startupReady != nil && h.startupReady.Load() {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{Status: "healthy", Gateway: "healthy"})
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(Response{Status: "starting", Gateway: "starting"})
}

// ServeReady handles GET /health/ready requests (readiness probe).
// Concurrently probes all configured upstream services and returns the
// aggregated result. Returns 200 if healthy/degraded, 503 if all unhealthy.
func (h *Handler) ServeReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	statuses := h.checkUpstreams(ctx)

	resp := Response{
		Gateway:  "healthy",
		Upstream: statuses,
	}

	unhealthyCount := 0
	for _, s := range statuses {
		if s.Status != "healthy" {
			unhealthyCount++
		}
	}

	switch {
	case unhealthyCount == 0:
		resp.Status = "healthy"
	case unhealthyCount < len(statuses):
		resp.Status = "degraded"
	default:
		resp.Status = "unhealthy"
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("failed to encode health response")
	}
}

// ServeDetailed handles GET /health requests with a full diagnostic report.
// Includes version, uptime, startup status, and all upstream + infrastructure checks.
func (h *Handler) ServeDetailed(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	statuses := h.checkUpstreams(ctx)

	// Check Redis (rate limiter)
	if h.redisCheck != nil {
		redisStatus := Status{Service: "redis"}
		if h.redisCheck() {
			redisStatus.Status = "healthy"
		} else {
			redisStatus.Status = "unhealthy"
		}
		statuses = append(statuses, redisStatus)
	}

	unhealthyCount := 0
	for _, s := range statuses {
		if s.Status != "healthy" {
			unhealthyCount++
		}
	}

	status := "healthy"
	switch {
	case unhealthyCount == 0:
		// healthy
	case unhealthyCount < len(statuses):
		status = "degraded"
	default:
		status = "unhealthy"
	}

	resp := DetailedResponse{
		Status:           status,
		Service:          "gateway",
		Version:          h.version,
		Uptime:           time.Since(h.startTime).Round(time.Second).String(),
		StartupCompleted: h.startupReady != nil && h.startupReady.Load(),
		Gateway:          "healthy",
		Upstream:         statuses,
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("failed to encode detailed health response")
	}
}

// checkUpstreams probes each configured service concurrently and returns
// a slice of Status results.
func (h *Handler) checkUpstreams(ctx context.Context) []Status {
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		statuses []Status
	)

	for name, svc := range h.services {
		wg.Add(1)
		go func(name string, svc config.ServiceEntry) {
			defer wg.Done()
			status := h.checkService(ctx, name, svc)
			mu.Lock()
			statuses = append(statuses, status)
			mu.Unlock()
		}(name, svc)
	}

	wg.Wait()
	return statuses
}

// checkService sends an HTTP GET to the service's health endpoint and
// reports whether the response indicates a healthy state.
func (h *Handler) checkService(ctx context.Context, name string, svc config.ServiceEntry) Status {
	start := time.Now()

	addr, err := h.resolver.Resolve(name)
	if err != nil {
		return Status{
			Service: name,
			Status:  "unhealthy",
			Error:   fmt.Sprintf("resolution failed: %v", err),
		}
	}

	healthPath := svc.HealthPath
	if healthPath == "" {
		healthPath = "/health"
	}

	healthURL := addr + healthPath
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		healthURL = "http://" + addr + healthPath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return Status{
			Service: name,
			Status:  "unhealthy",
			Error:   fmt.Sprintf("creating request: %v", err),
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return Status{
			Service: name,
			Status:  "unhealthy",
			Latency: time.Since(start).String(),
			Error:   fmt.Sprintf("request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return Status{
			Service: name,
			Status:  "healthy",
			Latency: latency.String(),
		}
	}

	return Status{
		Service: name,
		Status:  "unhealthy",
		Latency: latency.String(),
		Error:   fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
	}
}
