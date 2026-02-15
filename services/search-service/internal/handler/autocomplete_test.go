package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/streamvault/search-service/internal/model"
)

func TestAutocompleteHandler_ValidQuery_Returns200(t *testing.T) {
	ms := &mockSearcher{
		autocompleteFn: func(_ context.Context, _ model.AutocompleteRequest) (*model.AutocompleteResponse, error) {
			return &model.AutocompleteResponse{
				Suggestions: []model.AutocompleteSuggestion{
					{ContentID: "c1", Title: "Interstellar", ContentType: "movie", ReleaseYear: 2014},
					{ContentID: "c2", Title: "Inception", ContentType: "movie", ReleaseYear: 2010},
				},
			}, nil
		},
	}
	h := NewAutocompleteHandler(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/search/autocomplete?q=Int", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp model.AutocompleteResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(resp.Suggestions))
	}
}

func TestAutocompleteHandler_ShortQuery_Returns400(t *testing.T) {
	h := NewAutocompleteHandler(&mockSearcher{})

	req := httptest.NewRequest(http.MethodGet, "/api/search/autocomplete?q=x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short query, got %d", rec.Code)
	}
}

func TestAutocompleteHandler_SearchError_Returns500(t *testing.T) {
	ms := &mockSearcher{
		autocompleteFn: func(_ context.Context, _ model.AutocompleteRequest) (*model.AutocompleteResponse, error) {
			return nil, errors.New("elasticsearch unavailable")
		},
	}
	h := NewAutocompleteHandler(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/search/autocomplete?q=test", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestAutocompleteHandler_EmptyResults_ReturnsEmptyArray(t *testing.T) {
	ms := &mockSearcher{
		autocompleteFn: func(_ context.Context, _ model.AutocompleteRequest) (*model.AutocompleteResponse, error) {
			return &model.AutocompleteResponse{
				Suggestions: []model.AutocompleteSuggestion{},
			}, nil
		},
	}
	h := NewAutocompleteHandler(ms)

	req := httptest.NewRequest(http.MethodGet, "/api/search/autocomplete?q=zzzzz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp model.AutocompleteResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(resp.Suggestions))
	}
}
