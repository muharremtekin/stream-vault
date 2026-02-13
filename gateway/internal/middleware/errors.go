package middleware

import (
	"encoding/json"
	"net/http"
	"time"
)

// ErrorResponse is the standard error body returned by the gateway.
type ErrorResponse struct {
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// WriteErrorResponse writes a JSON error body with the unified format.
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := ErrorResponse{
		Status:    statusCode,
		Error:     http.StatusText(statusCode),
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	_ = json.NewEncoder(w).Encode(resp)
}
