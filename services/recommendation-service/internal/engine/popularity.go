package engine

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
)

const (
	// Popularity score component weights.
	ratingWeight  = 0.4
	viewWeight    = 0.3
	recencyWeight = 0.3
	// recencyRange is the number of years used for recency normalization.
	recencyRange = 30
)

// PopularityEngine provides popularity-based recommendations.
// Used as a fallback for cold-start users (< 5 interactions).
type PopularityEngine struct {
	repo *repository.Repository
}

// NewPopularityEngine creates a new PopularityEngine.
func NewPopularityEngine(repo *repository.Repository) *PopularityEngine {
	return &PopularityEngine{repo: repo}
}

// Recommend returns popular content ranked by a composite popularity score.
// score = 0.4 * normalizedRating + 0.3 * normalizedViewCount + 0.3 * normalizedRecency
func (e *PopularityEngine) Recommend(ctx context.Context, userID string, limit int) ([]scoredItem, error) {
	allContent, err := e.repo.GetAllContentFeatures(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading content features: %w", err)
	}
	if len(allContent) == 0 {
		return nil, nil
	}

	viewCounts, err := e.repo.GetContentViewCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading view counts: %w", err)
	}

	// Find max view count for normalization
	maxViews := 1
	for _, cnt := range viewCounts {
		if cnt > maxViews {
			maxViews = cnt
		}
	}

	// Build set of fully watched content for exclusion
	watchedSet := make(map[string]bool)
	if userID != "" {
		interactions, err := e.repo.GetUserInteractions(ctx, userID, 1000)
		if err == nil {
			for _, i := range interactions {
				if i.InteractionType == model.InteractionComplete || i.CompletionPct >= 90 {
					watchedSet[i.ContentID] = true
				}
			}
		}
	}

	currentYear := time.Now().Year()

	var scored []scoredItem
	for _, cf := range allContent {
		if watchedSet[cf.ContentID] {
			continue
		}

		// Normalized rating [0,1]
		normRating := cf.AvgRating / 10.0
		if normRating > 1 {
			normRating = 1
		}

		// Normalized view count [0,1]
		normViews := float64(viewCounts[cf.ContentID]) / float64(maxViews)

		// Normalized recency [0,1]
		yearDiff := currentYear - cf.ReleaseYear
		normRecency := 1.0 - float64(yearDiff)/float64(recencyRange)
		if normRecency < 0 {
			normRecency = 0
		}
		if normRecency > 1 {
			normRecency = 1
		}

		score := ratingWeight*normRating + viewWeight*normViews + recencyWeight*normRecency
		scored = append(scored, scoredItem{ContentID: cf.ContentID, Score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if limit > 0 && len(scored) > limit {
		scored = scored[:limit]
	}
	return scored, nil
}
