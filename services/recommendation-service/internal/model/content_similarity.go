package model

import "time"

// ContentSimilarity stores the precomputed similarity score between two content items.
type ContentSimilarity struct {
	ContentIDA      string    `json:"contentIdA"`
	ContentIDB      string    `json:"contentIdB"`
	SimilarityScore float64   `json:"similarityScore"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
