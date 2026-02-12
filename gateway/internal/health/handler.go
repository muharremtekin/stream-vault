package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
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
	Upstream []Status `json:"upstream"`
}

// Handler serves the /health endpoint. It probes every configured downstream
// service and aggregates the results into a single response.
type Handler struct {
	services map[string]config.ServiceEntry
	resolver *discovery.Resolver
	client   *http.Client
}

// NewHandler creates a health Handler that checks the provided services.
func NewHandler(services map[string]config.ServiceEntry, resolver *discovery.Resolver) *Handler {
	return &Handler{
		services: services,
		resolver: resolver,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ServeHTTP handles GET /health requests. It concurrently probes all
// configured upstream services and returns the aggregated result.
// Returns 200 if all services are healthy, 503 if any are unhealthy.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
