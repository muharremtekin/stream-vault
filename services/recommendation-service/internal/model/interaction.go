package model

import "time"

// InteractionType enumerates user interaction types.
type InteractionType string

const (
	InteractionView     InteractionType = "view"
	InteractionComplete InteractionType = "complete"
	InteractionRate     InteractionType = "rate"
	InteractionBookmark InteractionType = "bookmark"
)

// Interaction records a user's interaction with a piece of content.
type Interaction struct {
	ID              int64           `json:"id"`
	UserID          string          `json:"userId"`
	ContentID       string          `json:"contentId"`
	InteractionType InteractionType `json:"interactionType"`
	Rating          *float64        `json:"rating,omitempty"`
	CompletionPct   float64         `json:"completionPct"`
	CreatedAt       time.Time       `json:"createdAt"`
}
