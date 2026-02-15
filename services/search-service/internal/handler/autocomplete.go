package handler

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/metrics"
	"github.com/streamvault/search-service/internal/model"
)

// AutocompleteHandler handles the autocomplete endpoint.
type AutocompleteHandler struct {
	searcher Searcher
}

// NewAutocompleteHandler creates a new AutocompleteHandler.
func NewAutocompleteHandler(searcher Searcher) *AutocompleteHandler {
	return &AutocompleteHandler{searcher: searcher}
}

// ServeHTTP handles GET /api/search/autocomplete requests.
func (h *AutocompleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := model.ParseAutocompleteRequest(r)

	if len(req.Query) < 2 {
		WriteErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "query must be at least 2 characters")
		return
	}

	metrics.SearchQueriesTotal.WithLabelValues("autocomplete").Inc()
	resp, err := h.searcher.Autocomplete(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Str("query", req.Query).Msg("autocomplete failed")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "SEARCH_ERROR", "autocomplete request failed")
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}
