package handler

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/elasticsearch"
	"github.com/streamvault/search-service/internal/model"
)

// SearchHandler handles the full-text search endpoint.
type SearchHandler struct {
	searcher *elasticsearch.Searcher
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(searcher *elasticsearch.Searcher) *SearchHandler {
	return &SearchHandler{searcher: searcher}
}

// ServeHTTP handles GET /api/search requests.
func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := model.ParseSearchRequest(r)

	// Validate pagination bounds
	var errors []FieldError
	if req.Page < 1 {
		errors = append(errors, FieldError{Field: "page", Message: "must be at least 1"})
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		errors = append(errors, FieldError{Field: "pageSize", Message: "must be between 1 and 100"})
	}
	if len(errors) > 0 {
		WriteValidationError(w, r, errors)
		return
	}

	resp, err := h.searcher.Search(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Str("query", req.Query).Msg("search failed")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "SEARCH_ERROR", "search request failed")
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}
