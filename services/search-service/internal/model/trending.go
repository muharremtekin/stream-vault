package model

import (
	"net/http"
	"strconv"
)

// TrendingRequest holds parsed query parameters for the trending endpoint.
type TrendingRequest struct {
	TimeWindow string
	Limit      int
}

// ParseTrendingRequest parses query parameters from an HTTP request.
func ParseTrendingRequest(r *http.Request) TrendingRequest {
	q := r.URL.Query()

	window := q.Get("window")
	if window == "" {
		window = "week"
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return TrendingRequest{
		TimeWindow: window,
		Limit:      limit,
	}
}

// TrendingResponse is the JSON response for the trending endpoint.
type TrendingResponse struct {
	Items []TrendingItem `json:"items"`
}

// TrendingItem represents a single trending content item.
type TrendingItem struct {
	ContentID     string  `json:"contentId"`
	Title         string  `json:"title"`
	ThumbnailURL  string  `json:"thumbnailUrl"`
	ContentType   string  `json:"contentType"`
	ReleaseYear   int     `json:"releaseYear"`
	AverageRating float64 `json:"averageRating"`
	Rank          int     `json:"rank"`
	RankChange    int     `json:"rankChange"`
	ViewCount     int64   `json:"viewCount"`
}
