package model

import (
	"encoding/json"
	"time"
)

// Event is the standard envelope for all RabbitMQ messages (Rule 3.4).
type Event struct {
	EventID       string          `json:"eventId"`
	EventType     string          `json:"eventType"`
	Timestamp     string          `json:"timestamp"`
	Source        string          `json:"source"`
	CorrelationID string          `json:"correlationId"`
	Data          json.RawMessage `json:"data"`
}

// ContentEventData carries content fields for created/updated events.
// JSON keys use camelCase matching the catalog service API output.
type ContentEventData struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	OriginalTitle   string   `json:"originalTitle,omitempty"`
	Description     string   `json:"description"`
	ContentType     string   `json:"contentType"`
	Genres          []string `json:"genres"`
	Tags            []string `json:"tags,omitempty"`
	CastNames       []string `json:"castNames,omitempty"`
	Director        string   `json:"director,omitempty"`
	ReleaseYear     int      `json:"releaseYear"`
	MaturityRating  string   `json:"maturityRating"`
	AverageRating   float64  `json:"averageRating"`
	RatingCount     int      `json:"ratingCount"`
	ThumbnailURL    string   `json:"thumbnailUrl,omitempty"`
	BannerURL       string   `json:"bannerUrl,omitempty"`
	DurationSeconds int      `json:"durationSeconds,omitempty"`
	SeasonCount     int      `json:"seasonCount,omitempty"`
	EpisodeCount    int      `json:"episodeCount,omitempty"`
}

// ToSearchDocument converts a ContentEventData to a SearchDocument for ES indexing.
func (d ContentEventData) ToSearchDocument() SearchDocument {
	now := time.Now().UTC()
	return SearchDocument{
		ID:              d.ID,
		Title:           d.Title,
		OriginalTitle:   d.OriginalTitle,
		Description:     d.Description,
		ContentType:     d.ContentType,
		Genres:          d.Genres,
		Tags:            d.Tags,
		CastNames:       d.CastNames,
		Director:        d.Director,
		ReleaseYear:     d.ReleaseYear,
		MaturityRating:  d.MaturityRating,
		AverageRating:   d.AverageRating,
		RatingCount:     d.RatingCount,
		ThumbnailURL:    d.ThumbnailURL,
		BannerURL:       d.BannerURL,
		DurationSeconds: d.DurationSeconds,
		SeasonCount:     d.SeasonCount,
		EpisodeCount:    d.EpisodeCount,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// DeleteEventData carries the content ID for deletion events.
type DeleteEventData struct {
	ID string `json:"id"`
}

// WatchEventData carries watch completion event data.
type WatchEventData struct {
	UserID    string `json:"userId"`
	ContentID string `json:"contentId"`
}
