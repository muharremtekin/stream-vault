package model

import (
	"math"
	"net/http"
	"strconv"
	"strings"
)

// SearchRequest holds parsed query parameters for the search endpoint.
type SearchRequest struct {
	Query           string
	Genres          []string
	YearFrom        int
	YearTo          int
	MaturityRatings []string
	MinRating       float64
	ContentType     string
	SortField       string
	SortOrder       string
	Page            int
	PageSize        int
}

// ParseSearchRequest parses query parameters from an HTTP request.
func ParseSearchRequest(r *http.Request) SearchRequest {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	yearFrom, _ := strconv.Atoi(q.Get("year_from"))
	yearTo, _ := strconv.Atoi(q.Get("year_to"))
	minRating, _ := strconv.ParseFloat(q.Get("min_rating"), 64)

	var genres []string
	if g := q.Get("genres"); g != "" {
		genres = strings.Split(g, ",")
	}

	var maturityRatings []string
	if m := q.Get("maturity_ratings"); m != "" {
		maturityRatings = strings.Split(m, ",")
	}

	sortField := q.Get("sort")
	if sortField == "" {
		sortField = "relevance"
	}

	sortOrder := q.Get("order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	return SearchRequest{
		Query:           q.Get("q"),
		Genres:          genres,
		YearFrom:        yearFrom,
		YearTo:          yearTo,
		MaturityRatings: maturityRatings,
		MinRating:       minRating,
		ContentType:     q.Get("content_type"),
		SortField:       sortField,
		SortOrder:       sortOrder,
		Page:            page,
		PageSize:        pageSize,
	}
}

// From returns the ES "from" offset for pagination.
func (r SearchRequest) From() int {
	return (r.Page - 1) * r.PageSize
}

// SearchResponse is the JSON response for the search endpoint.
type SearchResponse struct {
	Items      []SearchHit `json:"items"`
	Facets     *Facets     `json:"facets,omitempty"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalCount int64       `json:"totalCount"`
	TotalPages int         `json:"totalPages"`
}

// NewSearchResponse creates a SearchResponse with computed totalPages.
func NewSearchResponse(items []SearchHit, facets *Facets, page, pageSize int, totalCount int64) SearchResponse {
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	if totalPages < 0 {
		totalPages = 0
	}
	return SearchResponse{
		Items:      items,
		Facets:     facets,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

// SearchHit represents a single search result.
type SearchHit struct {
	ContentID      string     `json:"contentId"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	ContentType    string     `json:"contentType"`
	ThumbnailURL   string     `json:"thumbnailUrl"`
	ReleaseYear    int        `json:"releaseYear"`
	AverageRating  float64    `json:"averageRating"`
	MaturityRating string     `json:"maturityRating"`
	Genres         []string   `json:"genres"`
	Highlights     *Highlight `json:"highlights,omitempty"`
}

// Highlight contains highlighted text fragments with <em> tags.
type Highlight struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Facets contains aggregation buckets for filtering.
type Facets struct {
	GenreFacets       []FacetBucket `json:"genreFacets"`
	YearFacets        []FacetBucket `json:"yearFacets"`
	MaturityFacets    []FacetBucket `json:"maturityFacets"`
	ContentTypeFacets []FacetBucket `json:"contentTypeFacets"`
}

// FacetBucket represents a single aggregation bucket.
type FacetBucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"docCount"`
}
