package resilience

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// degradedResponse is the JSON envelope returned when a service is
// unavailable and the gateway provides a fallback.
type degradedResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// FallbackStrategy maps a service name to an HTTP handler that returns a
// graceful degradation response when the circuit breaker is open.
// Services without an entry get a generic 503.
var fallbackStrategies = map[string]func(w http.ResponseWriter, r *http.Request){
	// Search down → return empty results so the UI doesn't break.
	"search-service": func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusServiceUnavailable, degradedResponse{
			Status:  "degraded",
			Message: "search service temporarily unavailable, please try again later",
			Data: map[string]interface{}{
				"items":      []interface{}{},
				"totalCount": 0,
			},
		})
	},

	// Recommendation down → return empty recommendations.
	"recommendation-service": func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusServiceUnavailable, degradedResponse{
			Status:  "degraded",
			Message: "recommendation service temporarily unavailable",
			Data: map[string]interface{}{
				"recommendations": []interface{}{},
			},
		})
	},

	// Notification down → core flow unaffected, just inform the client.
	"notification-service": func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusServiceUnavailable, degradedResponse{
			Status:  "degraded",
			Message: "notification service temporarily unavailable, notifications may be delayed",
		})
	},

	// Encoding down → uploads are accepted by streaming service and queued;
	// inform the client that processing is delayed.
	"encoding-service": func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusServiceUnavailable, degradedResponse{
			Status:  "degraded",
			Message: "encoding service temporarily unavailable, jobs will be processed when service recovers",
		})
	},
}

// ServeFallback writes a graceful degradation response for the given service.
// Returns true if a fallback was served, false if no strategy exists (caller
// should write a generic 503).
func ServeFallback(w http.ResponseWriter, r *http.Request, service string) bool {
	fn, ok := fallbackStrategies[service]
	if !ok {
		return false
	}
	log.Info().Str("service", service).Msg("serving degraded fallback response")
	fn(w, r)
	return true
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("failed to encode fallback response")
	}
}
