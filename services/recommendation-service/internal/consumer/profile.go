package consumer

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/cache"
	"github.com/streamvault/recommendation-service/internal/engine"
	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
)

// refreshUserProfile recomputes and upserts a user's recommendation profile
// based on their interaction history. Called after watch/rating events.
func refreshUserProfile(ctx context.Context, repo *repository.Repository, featureCache *cache.Cache, userID string) {
	interactions, err := repo.GetUserInteractions(ctx, userID, 100)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to load interactions for profile refresh")
		return
	}
	if len(interactions) == 0 {
		return
	}

	allContent, err := repo.GetAllContentFeatures(ctx)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to load content features for profile refresh")
		return
	}

	contentMap := make(map[string]*model.ContentFeatures, len(allContent))
	for idx := range allContent {
		contentMap[allContent[idx].ContentID] = &allContent[idx]
	}

	// Compute genre weights
	genreWeights := make(map[string]float64)
	var totalWeight float64
	var ratingSum float64
	var ratingCount int

	for idx := range interactions {
		i := &interactions[idx]
		cf := contentMap[i.ContentID]
		if cf == nil {
			continue
		}

		weight := interactionWeight(i)
		for _, genre := range cf.Genres {
			genreWeights[genre] += weight
		}
		totalWeight += weight

		if i.InteractionType == model.InteractionRate && i.Rating != nil {
			ratingSum += *i.Rating
			ratingCount++
		}
	}

	// Normalize genre weights
	if totalWeight > 0 {
		for genre := range genreWeights {
			genreWeights[genre] /= totalWeight
		}
	}

	var avgRating float64
	if ratingCount > 0 {
		avgRating = ratingSum / float64(ratingCount)
	}

	profile := &model.UserProfile{
		UserID:            userID,
		GenreWeights:      genreWeights,
		AvgRating:         avgRating,
		TotalInteractions: len(interactions),
		UpdatedAt:         time.Now(),
	}

	if err := repo.UpsertUserProfile(ctx, profile); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to upsert user profile")
		return
	}

	featureCache.InvalidateUserProfile(ctx, userID)
	log.Debug().Str("user_id", userID).Int("interactions", len(interactions)).Msg("user profile refreshed")
}

// interactionWeight returns the weight for an interaction, matching the engine's logic.
func interactionWeight(i *model.Interaction) float64 {
	switch i.InteractionType {
	case model.InteractionComplete:
		return 1.0
	case model.InteractionRate:
		if i.Rating != nil {
			return *i.Rating / 10.0
		}
		return 0.5
	case model.InteractionBookmark:
		return 0.6
	case model.InteractionView:
		if i.CompletionPct >= 50 {
			return 0.5
		}
		return 0.2
	default:
		return 0.2
	}
}

// buildContentFeaturesFromEvent converts a catalog event into a ContentFeatures model
// with a computed feature vector.
func buildContentFeaturesFromEvent(data *model.ContentEventData) *model.ContentFeatures {
	cf := &model.ContentFeatures{
		ContentID:   data.ID,
		Title:       data.Title,
		Genres:      data.Genres,
		Tags:        data.Tags,
		Director:    data.Director,
		ReleaseYear: data.ReleaseYear,
		AvgRating:   data.AverageRating,
		UpdatedAt:   time.Now(),
	}
	cf.FeatureVector = engine.BuildFeatureVector(cf)
	return cf
}
