package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestCorrelationID_GeneratesWhenMissing(t *testing.T) {
	var downstreamCID string

	handler := CorrelationID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamCID = r.Header.Get("X-Correlation-Id")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	responseCID := rec.Header().Get("X-Correlation-Id")
	if responseCID == "" {
		t.Fatal("expected X-Correlation-Id on response, got empty")
	}
	if !uuidRe.MatchString(responseCID) {
		t.Errorf("expected UUID format, got %q", responseCID)
	}
	if downstreamCID != responseCID {
		t.Errorf("downstream request CID %q != response CID %q", downstreamCID, responseCID)
	}
}

func TestCorrelationID_PropagatesExisting(t *testing.T) {
	const existingID = "abc-123-existing"
	var downstreamCID string

	handler := CorrelationID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamCID = r.Header.Get("X-Correlation-Id")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-Id", existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	responseCID := rec.Header().Get("X-Correlation-Id")
	if responseCID != existingID {
		t.Errorf("expected response CID %q, got %q", existingID, responseCID)
	}
	if downstreamCID != existingID {
		t.Errorf("expected downstream CID %q, got %q", existingID, downstreamCID)
	}
}

func TestCorrelationID_SetOnBothRequestAndResponse(t *testing.T) {
	var requestCID string

	handler := CorrelationID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCID = r.Header.Get("X-Correlation-Id")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	responseCID := rec.Header().Get("X-Correlation-Id")
	if requestCID == "" {
		t.Error("expected X-Correlation-Id on downstream request")
	}
	if responseCID == "" {
		t.Error("expected X-Correlation-Id on response")
	}
	if requestCID != responseCID {
		t.Errorf("request CID %q != response CID %q", requestCID, responseCID)
	}
}
