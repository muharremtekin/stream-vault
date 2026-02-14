package engine

import "context"

// Recommender is the central abstraction consumed by both HTTP and gRPC layers.
type Recommender interface {
	// GetRecommendations returns personalized recommendations for a user.
	GetRecommendations(ctx context.Context, userID string, limit int) ([]RecommendedItem, error)

	// GetSimilar returns content similar to the given content ID.
	GetSimilar(ctx context.Context, contentID string, limit int) ([]SimilarItem, error)

	// GetHomePageSections returns curated home page sections for a user.
	GetHomePageSections(ctx context.Context, userID string) ([]HomePageSection, error)

	// RecordFeedback records user feedback (e.g., not_interested).
	RecordFeedback(ctx context.Context, userID string, contentID string, feedbackType string) error
}

// RecommendedItem is a single recommendation result.
type RecommendedItem struct {
	ContentID     string  `json:"contentId"`
	Title         string  `json:"title"`
	ThumbnailURL  string  `json:"thumbnailUrl,omitempty"`
	ContentType   string  `json:"contentType"`
	ReleaseYear   int     `json:"releaseYear"`
	AverageRating float64 `json:"averageRating"`
	Score         float64 `json:"score"`
	Algorithm     string  `json:"algorithm"`
	Reason        string  `json:"reason"`
}

// SimilarItem is a single similar-content result.
type SimilarItem struct {
	ContentID       string  `json:"contentId"`
	Title           string  `json:"title"`
	ThumbnailURL    string  `json:"thumbnailUrl,omitempty"`
	ContentType     string  `json:"contentType"`
	ReleaseYear     int     `json:"releaseYear"`
	AverageRating   float64 `json:"averageRating"`
	SimilarityScore float64 `json:"similarityScore"`
}

// HomePageSection represents one section on the home page.
type HomePageSection struct {
	SectionType string            `json:"sectionType"`
	Title       string            `json:"title"`
	Items       []RecommendedItem `json:"items"`
}

// scoredItem is an internal type used between sub-engines.
type scoredItem struct {
	ContentID string
	Score     float64
}
