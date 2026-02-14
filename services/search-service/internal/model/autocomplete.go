package model

import (
	"net/http"
	"strconv"
)

// AutocompleteRequest holds parsed query parameters for autocomplete.
type AutocompleteRequest struct {
	Query string
	Limit int
}

// ParseAutocompleteRequest parses query parameters from an HTTP request.
func ParseAutocompleteRequest(r *http.Request) AutocompleteRequest {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 20 {
		limit = 5
	}

	return AutocompleteRequest{
		Query: q.Get("q"),
		Limit: limit,
	}
}

// AutocompleteResponse is the JSON response for the autocomplete endpoint.
type AutocompleteResponse struct {
	Suggestions []AutocompleteSuggestion `json:"suggestions"`
}

// AutocompleteSuggestion represents a single autocomplete suggestion.
type AutocompleteSuggestion struct {
	ContentID    string `json:"contentId"`
	Title        string `json:"title"`
	ContentType  string `json:"contentType"`
	ThumbnailURL string `json:"thumbnailUrl"`
	ReleaseYear  int    `json:"releaseYear"`
}
