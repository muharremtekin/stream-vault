package model

import "time"

// UserProfile stores aggregated preferences for a user.
type UserProfile struct {
	UserID            string             `json:"userId"`
	GenreWeights      map[string]float64 `json:"genreWeights"`
	AvgRating         float64            `json:"avgRating"`
	TotalInteractions int                `json:"totalInteractions"`
	UpdatedAt         time.Time          `json:"updatedAt"`
}
