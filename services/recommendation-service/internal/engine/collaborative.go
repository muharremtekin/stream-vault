package engine

import (
	"context"
	"fmt"
	"sort"

	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
)

const (
	kNeighbors = 20
)

// CollaborativeEngine provides user-based collaborative filtering.
type CollaborativeEngine struct {
	repo *repository.Repository
}

// NewCollaborativeEngine creates a new CollaborativeEngine.
func NewCollaborativeEngine(repo *repository.Repository) *CollaborativeEngine {
	return &CollaborativeEngine{repo: repo}
}

// userSimilarity holds a user ID and their similarity score to the target user.
type userSimilarity struct {
	UserID     string
	Similarity float64
}

// implicitRating converts an interaction to an implicit rating value (1-10 scale).
func implicitRating(interaction *model.Interaction) float64 {
	switch interaction.InteractionType {
	case model.InteractionRate:
		if interaction.Rating != nil {
			return *interaction.Rating
		}
		return 5.0
	case model.InteractionComplete:
		return 7.0
	case model.InteractionBookmark:
		return 6.0
	case model.InteractionView:
		if interaction.CompletionPct >= 90 {
			return 7.0
		}
		if interaction.CompletionPct >= 50 {
			return 5.0
		}
		return 3.0
	default:
		return 3.0
	}
}

// buildUserItemMatrix loads all interactions and builds a user-item rating matrix.
// For duplicate (user, content) pairs, the highest implicit rating wins.
func (e *CollaborativeEngine) buildUserItemMatrix(ctx context.Context) (map[string]map[string]float64, error) {
	interactions, err := e.repo.GetAllInteractions(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading all interactions: %w", err)
	}

	matrix := make(map[string]map[string]float64)
	for idx := range interactions {
		interaction := &interactions[idx]
		rating := implicitRating(interaction)

		userVec, ok := matrix[interaction.UserID]
		if !ok {
			userVec = make(map[string]float64)
			matrix[interaction.UserID] = userVec
		}

		if existing, exists := userVec[interaction.ContentID]; !exists || rating > existing {
			userVec[interaction.ContentID] = rating
		}
	}
	return matrix, nil
}

// findSimilarUsers computes cosine similarity between the target user's
// rating vector and all other users' vectors. Returns the top K most
// similar users with positive similarity.
func findSimilarUsers(
	targetUserID string,
	userItemMatrix map[string]map[string]float64,
	k int,
) []userSimilarity {
	targetVec, ok := userItemMatrix[targetUserID]
	if !ok || len(targetVec) == 0 {
		return nil
	}

	var similarities []userSimilarity
	for userID, userVec := range userItemMatrix {
		if userID == targetUserID {
			continue
		}
		sim := CosineSimilarity(targetVec, userVec)
		if sim > 0 {
			similarities = append(similarities, userSimilarity{
				UserID:     userID,
				Similarity: sim,
			})
		}
	}

	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	if len(similarities) > k {
		similarities = similarities[:k]
	}
	return similarities
}

// predictRating predicts a rating for a content item using weighted average
// of similar users' ratings.
func predictRating(
	contentID string,
	similarUsers []userSimilarity,
	userItemMatrix map[string]map[string]float64,
) float64 {
	var weightedSum, simSum float64

	for _, su := range similarUsers {
		userVec := userItemMatrix[su.UserID]
		rating, ok := userVec[contentID]
		if !ok {
			continue
		}
		weightedSum += su.Similarity * rating
		simSum += su.Similarity
	}

	if simSum == 0 {
		return 0
	}
	return weightedSum / simSum
}

// Recommend returns collaborative filtering recommendations for a user.
func (e *CollaborativeEngine) Recommend(ctx context.Context, userID string, limit int) ([]scoredItem, error) {
	// Build user-item matrix
	matrix, err := e.buildUserItemMatrix(ctx)
	if err != nil {
		return nil, err
	}

	if len(matrix) < 2 {
		return nil, nil // Need at least 2 users for collaborative filtering
	}

	// Find similar users
	similarUsers := findSimilarUsers(userID, matrix, kNeighbors)
	if len(similarUsers) == 0 {
		return nil, nil
	}

	// Get set of content the user has already interacted with
	userRated := matrix[userID]

	// Collect all candidate content IDs from similar users
	candidates := make(map[string]bool)
	for _, su := range similarUsers {
		for contentID := range matrix[su.UserID] {
			if _, seen := userRated[contentID]; !seen {
				candidates[contentID] = true
			}
		}
	}

	// Predict ratings for each candidate
	var scored []scoredItem
	for contentID := range candidates {
		predicted := predictRating(contentID, similarUsers, matrix)
		if predicted > 0 {
			// Normalize to [0,1] range (predicted is on 1-10 scale)
			normalizedScore := predicted / 10.0
			if normalizedScore > 1 {
				normalizedScore = 1
			}
			scored = append(scored, scoredItem{ContentID: contentID, Score: normalizedScore})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if limit > 0 && len(scored) > limit {
		scored = scored[:limit]
	}
	return scored, nil
}
