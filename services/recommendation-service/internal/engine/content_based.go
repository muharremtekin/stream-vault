package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
)

const (
	// minYear is the baseline year for normalizing release years.
	minYear = 1950
	// topNSimilar is the number of similar items to store per content in the similarity matrix.
	topNSimilar = 10
)

// ContentBasedEngine provides content-based filtering recommendations.
type ContentBasedEngine struct {
	repo *repository.Repository
}

// NewContentBasedEngine creates a new ContentBasedEngine.
func NewContentBasedEngine(repo *repository.Repository) *ContentBasedEngine {
	return &ContentBasedEngine{repo: repo}
}

// BuildFeatureVector constructs a normalized feature vector for a ContentFeatures item.
// Keys follow naming conventions:
//   - "genre:<lowercase>" → 1.0 per genre
//   - "tag:<lowercase-hyphenated>" → 1.0 per tag
//   - "director:<lowercase-hyphenated>" → 1.0
//   - "year" → normalized to [0,1]: (releaseYear - 1950) / (currentYear - 1950)
//   - "rating" → avgRating / 10.0
func BuildFeatureVector(cf *model.ContentFeatures) map[string]float64 {
	vec := make(map[string]float64)

	for _, genre := range cf.Genres {
		vec[normalizeKey("genre", genre)] = 1.0
	}

	for _, tag := range cf.Tags {
		vec[normalizeKey("tag", tag)] = 1.0
	}

	if cf.Director != "" {
		vec[normalizeKey("director", cf.Director)] = 1.0
	}

	currentYear := time.Now().Year()
	yearRange := float64(currentYear - minYear)
	if yearRange > 0 {
		yearNorm := float64(cf.ReleaseYear-minYear) / yearRange
		if yearNorm < 0 {
			yearNorm = 0
		}
		if yearNorm > 1 {
			yearNorm = 1
		}
		vec["year"] = yearNorm
	}

	if cf.AvgRating > 0 {
		vec["rating"] = cf.AvgRating / 10.0
	}

	return vec
}

// normalizeKey creates a feature key from a prefix and value.
func normalizeKey(prefix, value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	return prefix + ":" + normalized
}

// interactionWeight returns the weight for an interaction used in building
// the user preference vector.
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

// BuildUserPreferenceVector creates a weighted average feature vector from
// a user's interactions. Each content's feature vector is weighted by interaction strength.
func BuildUserPreferenceVector(
	interactions []model.Interaction,
	contentMap map[string]*model.ContentFeatures,
) map[string]float64 {
	if len(interactions) == 0 {
		return nil
	}

	prefVec := make(map[string]float64)
	var totalWeight float64

	for idx := range interactions {
		interaction := &interactions[idx]
		cf, ok := contentMap[interaction.ContentID]
		if !ok || cf == nil {
			continue
		}

		weight := interactionWeight(interaction)
		fv := cf.FeatureVector
		if len(fv) == 0 {
			fv = BuildFeatureVector(cf)
		}

		for key, val := range fv {
			prefVec[key] += val * weight
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return nil
	}

	// Normalize by total weight
	for key := range prefVec {
		prefVec[key] /= totalWeight
	}
	return prefVec
}

// ScoreContent scores a single content item against a user preference vector.
func ScoreContent(userPrefVector map[string]float64, cf *model.ContentFeatures) float64 {
	fv := cf.FeatureVector
	if len(fv) == 0 {
		fv = BuildFeatureVector(cf)
	}
	return CosineSimilarity(userPrefVector, fv)
}

// Recommend returns content-based recommendations for a user.
func (e *ContentBasedEngine) Recommend(ctx context.Context, userID string, limit int) ([]scoredItem, error) {
	// Load user's recent interactions
	interactions, err := e.repo.GetUserInteractions(ctx, userID, 100)
	if err != nil {
		return nil, fmt.Errorf("loading user interactions: %w", err)
	}
	if len(interactions) == 0 {
		return nil, nil
	}

	// Load all content features
	allContent, err := e.repo.GetAllContentFeatures(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading content features: %w", err)
	}

	// Build content map
	contentMap := make(map[string]*model.ContentFeatures, len(allContent))
	for idx := range allContent {
		contentMap[allContent[idx].ContentID] = &allContent[idx]
	}

	// Build user preference vector
	userPrefVec := BuildUserPreferenceVector(interactions, contentMap)
	if userPrefVec == nil {
		return nil, nil
	}

	// Build set of seen content
	seenContent := make(map[string]bool)
	for _, interaction := range interactions {
		if interaction.InteractionType == model.InteractionComplete ||
			interaction.CompletionPct >= 90 {
			seenContent[interaction.ContentID] = true
		}
	}

	// Score each unseen content
	var scored []scoredItem
	for _, cf := range allContent {
		if seenContent[cf.ContentID] {
			continue
		}
		score := ScoreContent(userPrefVec, &cf)
		if score > 0 {
			scored = append(scored, scoredItem{ContentID: cf.ContentID, Score: score})
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

// ComputeSimilarityMatrix computes content-content similarity for all content
// and stores the top-N pairs per content in the database.
func (e *ContentBasedEngine) ComputeSimilarityMatrix(ctx context.Context) error {
	allContent, err := e.repo.GetAllContentFeatures(ctx)
	if err != nil {
		return fmt.Errorf("loading content features: %w", err)
	}

	if len(allContent) < 2 {
		return nil
	}

	// Ensure all content has feature vectors
	for idx := range allContent {
		if len(allContent[idx].FeatureVector) == 0 {
			allContent[idx].FeatureVector = BuildFeatureVector(&allContent[idx])
		}
	}

	var allSims []model.ContentSimilarity

	for i := 0; i < len(allContent); i++ {
		var sims []struct {
			otherID string
			score   float64
		}

		for j := 0; j < len(allContent); j++ {
			if i == j {
				continue
			}
			score := CosineSimilarity(allContent[i].FeatureVector, allContent[j].FeatureVector)
			if score > 0.01 { // Skip near-zero similarities
				sims = append(sims, struct {
					otherID string
					score   float64
				}{allContent[j].ContentID, score})
			}
		}

		// Sort and keep top-N
		sort.Slice(sims, func(a, b int) bool {
			return sims[a].score > sims[b].score
		})
		if len(sims) > topNSimilar {
			sims = sims[:topNSimilar]
		}

		now := time.Now()
		for _, s := range sims {
			allSims = append(allSims, model.ContentSimilarity{
				ContentIDA:      allContent[i].ContentID,
				ContentIDB:      s.otherID,
				SimilarityScore: s.score,
				UpdatedAt:       now,
			})
		}
	}

	if len(allSims) > 0 {
		if err := e.repo.BatchUpsertContentSimilarity(ctx, allSims); err != nil {
			return fmt.Errorf("batch upserting similarity matrix: %w", err)
		}
		log.Info().Int("pairs", len(allSims)).Msg("content similarity matrix updated")
	}
	return nil
}

// GetSimilar returns content similar to the given content ID.
// Uses precomputed similarity if available, falls back to on-the-fly computation.
func (e *ContentBasedEngine) GetSimilar(
	ctx context.Context, contentID string, limit int,
	allContent []model.ContentFeatures,
) ([]scoredItem, error) {
	// Try precomputed similarities first
	precomputed, err := e.repo.GetSimilarContent(ctx, contentID, limit)
	if err != nil {
		log.Warn().Err(err).Str("content_id", contentID).Msg("failed to get precomputed similarities")
	}

	if len(precomputed) > 0 {
		var result []scoredItem
		for _, sim := range precomputed {
			otherID := sim.ContentIDB
			if otherID == contentID {
				otherID = sim.ContentIDA
			}
			result = append(result, scoredItem{ContentID: otherID, Score: sim.SimilarityScore})
		}
		return result, nil
	}

	// Fallback: on-the-fly computation
	if allContent == nil {
		allContent, err = e.repo.GetAllContentFeatures(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading content features: %w", err)
		}
	}

	// Find the target content
	var targetVec map[string]float64
	for _, cf := range allContent {
		if cf.ContentID == contentID {
			targetVec = cf.FeatureVector
			if len(targetVec) == 0 {
				targetVec = BuildFeatureVector(&cf)
			}
			break
		}
	}

	if targetVec == nil {
		return nil, nil
	}

	var scored []scoredItem
	for _, cf := range allContent {
		if cf.ContentID == contentID {
			continue
		}
		fv := cf.FeatureVector
		if len(fv) == 0 {
			fv = BuildFeatureVector(&cf)
		}
		score := CosineSimilarity(targetVec, fv)
		if score > 0 {
			scored = append(scored, scoredItem{ContentID: cf.ContentID, Score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if limit > 0 && len(scored) > limit {
		scored = scored[:limit]
	}
	return scored, nil
}
