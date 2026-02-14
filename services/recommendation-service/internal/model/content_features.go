package model

import "time"

// ContentFeatures stores the feature representation of content for recommendation algorithms.
type ContentFeatures struct {
	ContentID     string             `json:"contentId"`
	Title         string             `json:"title"`
	Genres        []string           `json:"genres"`
	Tags          []string           `json:"tags"`
	Director      string             `json:"director"`
	ReleaseYear   int                `json:"releaseYear"`
	AvgRating     float64            `json:"avgRating"`
	FeatureVector map[string]float64 `json:"featureVector"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}
