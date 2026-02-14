package model

import "time"

// SearchDocument represents a document stored in Elasticsearch.
// Field names match the streamvault-content index mapping.
type SearchDocument struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	OriginalTitle   string    `json:"original_title,omitempty"`
	Description     string    `json:"description"`
	ContentType     string    `json:"content_type"`
	Genres          []string  `json:"genres"`
	Tags            []string  `json:"tags,omitempty"`
	CastNames       []string  `json:"cast_names,omitempty"`
	Director        string    `json:"director,omitempty"`
	ReleaseYear     int       `json:"release_year"`
	MaturityRating  string    `json:"maturity_rating"`
	AverageRating   float64   `json:"average_rating"`
	RatingCount     int       `json:"rating_count"`
	ViewCount       int64     `json:"view_count"`
	VideoStatus     string    `json:"video_status"`
	ThumbnailURL    string    `json:"thumbnail_url,omitempty"`
	BannerURL       string    `json:"banner_url,omitempty"`
	DurationSeconds int       `json:"duration_seconds,omitempty"`
	SeasonCount     int       `json:"season_count,omitempty"`
	EpisodeCount    int       `json:"episode_count,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
