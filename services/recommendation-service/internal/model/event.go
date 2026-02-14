package model

import "encoding/json"

// Event is the standard envelope for all RabbitMQ messages.
type Event struct {
	EventID       string          `json:"eventId"`
	EventType     string          `json:"eventType"`
	Timestamp     string          `json:"timestamp"`
	Source        string          `json:"source"`
	CorrelationID string          `json:"correlationId"`
	Data          json.RawMessage `json:"data"`
}

// ContentEventData carries content fields from catalog events.
type ContentEventData struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	ContentType   string   `json:"contentType"`
	Genres        []string `json:"genres"`
	Tags          []string `json:"tags,omitempty"`
	Director      string   `json:"director,omitempty"`
	ReleaseYear   int      `json:"releaseYear"`
	AverageRating float64  `json:"averageRating"`
	ThumbnailURL  string   `json:"thumbnailUrl,omitempty"`
}

// WatchEventData carries watch completion event data.
type WatchEventData struct {
	UserID        string  `json:"userId"`
	ContentID     string  `json:"contentId"`
	CompletionPct float64 `json:"completionPct,omitempty"`
}

// RatingEventData carries user rating data.
type RatingEventData struct {
	UserID    string  `json:"userId"`
	ContentID string  `json:"contentId"`
	Rating    float64 `json:"rating"`
}
