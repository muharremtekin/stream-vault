package batch

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/cache"
	"github.com/streamvault/recommendation-service/internal/engine"
	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
)

// Scheduler runs periodic batch computations for the recommendation engine:
// - Refreshing content feature vectors
// - Recomputing the content similarity matrix
// - Refreshing all user profiles
type Scheduler struct {
	repo         *repository.Repository
	cache        *cache.Cache
	contentBased *engine.ContentBasedEngine
	interval     time.Duration
}

// NewScheduler creates a new batch computation scheduler.
func NewScheduler(repo *repository.Repository, cache *cache.Cache, contentBased *engine.ContentBasedEngine, interval time.Duration) *Scheduler {
	return &Scheduler{
		repo:         repo,
		cache:        cache,
		contentBased: contentBased,
		interval:     interval,
	}
}

// Start runs the batch scheduler. It performs an initial run after a short delay,
// then runs periodically at the configured interval.
func (s *Scheduler) Start(ctx context.Context) {
	log.Info().Dur("interval", s.interval).Msg("batch scheduler started")

	// Initial delay to let the service stabilize
	select {
	case <-ctx.Done():
		log.Info().Msg("batch scheduler stopped before initial run")
		return
	case <-time.After(1 * time.Minute):
	}

	s.runBatch(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("batch scheduler stopped")
			return
		case <-ticker.C:
			s.runBatch(ctx)
		}
	}
}

func (s *Scheduler) runBatch(ctx context.Context) {
	start := time.Now()
	log.Info().Msg("batch computation starting")

	featuresUpdated := s.refreshContentFeatures(ctx)
	s.computeSimilarityMatrix(ctx)
	profilesUpdated := s.refreshAllUserProfiles(ctx)

	log.Info().
		Int("features_updated", featuresUpdated).
		Int("profiles_updated", profilesUpdated).
		Dur("duration", time.Since(start)).
		Msg("batch computation completed")
}

// refreshContentFeatures recomputes feature vectors for all content and
// updates those that have changed.
func (s *Scheduler) refreshContentFeatures(ctx context.Context) int {
	allContent, err := s.repo.GetAllContentFeatures(ctx)
	if err != nil {
		log.Error().Err(err).Msg("batch: failed to load content features")
		return 0
	}

	updated := 0
	for idx := range allContent {
		cf := &allContent[idx]
		newVector := engine.BuildFeatureVector(cf)

		if !vectorsEqual(cf.FeatureVector, newVector) {
			cf.FeatureVector = newVector
			if err := s.repo.UpsertContentFeatures(ctx, cf); err != nil {
				log.Error().Err(err).Str("content_id", cf.ContentID).Msg("batch: failed to update content features")
				continue
			}
			s.cache.InvalidateContentFeatures(ctx, cf.ContentID)
			updated++
		}
	}

	if updated > 0 {
		log.Info().Int("count", updated).Msg("batch: content feature vectors updated")
	}
	return updated
}

// computeSimilarityMatrix delegates to the existing engine method.
func (s *Scheduler) computeSimilarityMatrix(ctx context.Context) {
	if err := s.contentBased.ComputeSimilarityMatrix(ctx); err != nil {
		log.Error().Err(err).Msg("batch: failed to compute similarity matrix")
	}
}

// refreshAllUserProfiles recomputes profiles for every user with interactions.
func (s *Scheduler) refreshAllUserProfiles(ctx context.Context) int {
	userIDs, err := s.repo.GetDistinctUserIDs(ctx)
	if err != nil {
		log.Error().Err(err).Msg("batch: failed to get distinct user IDs")
		return 0
	}

	if len(userIDs) == 0 {
		return 0
	}

	// Load content features once for all users
	allContent, err := s.repo.GetAllContentFeatures(ctx)
	if err != nil {
		log.Error().Err(err).Msg("batch: failed to load content features for profile refresh")
		return 0
	}

	contentMap := make(map[string]*model.ContentFeatures, len(allContent))
	for idx := range allContent {
		contentMap[allContent[idx].ContentID] = &allContent[idx]
	}

	updated := 0
	for _, userID := range userIDs {
		if err := s.refreshSingleUserProfile(ctx, userID, contentMap); err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("batch: failed to refresh user profile")
			continue
		}
		s.cache.InvalidateUserProfile(ctx, userID)
		s.cache.InvalidateRecommendations(ctx, userID)
		updated++
	}

	if updated > 0 {
		log.Info().Int("count", updated).Msg("batch: user profiles refreshed")
	}
	return updated
}

func (s *Scheduler) refreshSingleUserProfile(ctx context.Context, userID string, contentMap map[string]*model.ContentFeatures) error {
	interactions, err := s.repo.GetUserInteractions(ctx, userID, 100)
	if err != nil {
		return err
	}
	if len(interactions) == 0 {
		return nil
	}

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

	return s.repo.UpsertUserProfile(ctx, profile)
}

// interactionWeight mirrors the engine's weighting logic.
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

// vectorsEqual checks if two sparse feature vectors are identical.
func vectorsEqual(a, b map[string]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || v != bv {
			return false
		}
	}
	return true
}
