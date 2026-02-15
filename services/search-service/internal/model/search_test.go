package model

import (
	"net/http/httptest"
	"testing"
)

func TestParseSearchRequest_DefaultValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/search", nil)
	sr := ParseSearchRequest(req)

	if sr.Page != 1 {
		t.Errorf("expected default page=1, got %d", sr.Page)
	}
	if sr.PageSize != 20 {
		t.Errorf("expected default pageSize=20, got %d", sr.PageSize)
	}
	if sr.SortField != "relevance" {
		t.Errorf("expected default sort=relevance, got %q", sr.SortField)
	}
	if sr.SortOrder != "desc" {
		t.Errorf("expected default order=desc, got %q", sr.SortOrder)
	}
	if sr.Query != "" {
		t.Errorf("expected empty query, got %q", sr.Query)
	}
}

func TestParseSearchRequest_AllParams(t *testing.T) {
	req := httptest.NewRequest("GET",
		"/api/search?q=matrix&genres=action,scifi&year_from=1999&year_to=2003&maturity_ratings=PG-13,R&min_rating=8.0&content_type=movie&sort=rating&order=asc&page=2&pageSize=10",
		nil)
	sr := ParseSearchRequest(req)

	if sr.Query != "matrix" {
		t.Errorf("expected query 'matrix', got %q", sr.Query)
	}
	if len(sr.Genres) != 2 || sr.Genres[0] != "action" || sr.Genres[1] != "scifi" {
		t.Errorf("expected genres [action, scifi], got %v", sr.Genres)
	}
	if sr.YearFrom != 1999 {
		t.Errorf("expected yearFrom 1999, got %d", sr.YearFrom)
	}
	if sr.YearTo != 2003 {
		t.Errorf("expected yearTo 2003, got %d", sr.YearTo)
	}
	if len(sr.MaturityRatings) != 2 {
		t.Errorf("expected 2 maturity ratings, got %d", len(sr.MaturityRatings))
	}
	if sr.MinRating != 8.0 {
		t.Errorf("expected minRating 8.0, got %f", sr.MinRating)
	}
	if sr.ContentType != "movie" {
		t.Errorf("expected contentType 'movie', got %q", sr.ContentType)
	}
	if sr.SortField != "rating" {
		t.Errorf("expected sort 'rating', got %q", sr.SortField)
	}
	if sr.SortOrder != "asc" {
		t.Errorf("expected order 'asc', got %q", sr.SortOrder)
	}
	if sr.Page != 2 {
		t.Errorf("expected page 2, got %d", sr.Page)
	}
	if sr.PageSize != 10 {
		t.Errorf("expected pageSize 10, got %d", sr.PageSize)
	}
}

func TestParseSearchRequest_GenresSplitByComma(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/search?genres=action,drama,comedy", nil)
	sr := ParseSearchRequest(req)

	if len(sr.Genres) != 3 {
		t.Fatalf("expected 3 genres, got %d", len(sr.Genres))
	}
	expected := []string{"action", "drama", "comedy"}
	for i, g := range expected {
		if sr.Genres[i] != g {
			t.Errorf("expected genre[%d]=%q, got %q", i, g, sr.Genres[i])
		}
	}
}

func TestSearchRequest_From(t *testing.T) {
	tests := []struct {
		page, pageSize, expected int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{5, 5, 20},
	}

	for _, tt := range tests {
		sr := SearchRequest{Page: tt.page, PageSize: tt.pageSize}
		got := sr.From()
		if got != tt.expected {
			t.Errorf("page=%d, pageSize=%d: expected From()=%d, got %d",
				tt.page, tt.pageSize, tt.expected, got)
		}
	}
}

func TestParseAutocompleteRequest_DefaultLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/search/autocomplete?q=test", nil)
	ar := ParseAutocompleteRequest(req)

	if ar.Limit != 5 {
		t.Errorf("expected default limit=5, got %d", ar.Limit)
	}
	if ar.Query != "test" {
		t.Errorf("expected query 'test', got %q", ar.Query)
	}
}
