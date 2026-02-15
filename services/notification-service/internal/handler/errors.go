package handler

import (
	"encoding/json"
	"net/http"
)

const correlationHeader = "X-Correlation-Id"

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details,omitempty"`
	TraceID string        `json:"traceId"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	traceID := r.Header.Get(correlationHeader)
	resp := errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
			TraceID: traceID,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func WriteValidationError(w http.ResponseWriter, r *http.Request, details []FieldError) {
	traceID := r.Header.Get(correlationHeader)
	detailsIface := make([]interface{}, len(details))
	for i, d := range details {
		detailsIface[i] = d
	}
	resp := errorResponse{
		Error: errorBody{
			Code:    "VALIDATION_ERROR",
			Message: "request validation failed",
			Details: detailsIface,
			TraceID: traceID,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(resp)
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
