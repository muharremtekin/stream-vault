package engine

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/cache"
	"github.com/streamvault/recommendation-service/internal/metrics"
	"github.com/streamvault/recommendation-service/internal/model"
	"github.com/streamvault/recommendation-service/internal/repository"
)

const (
	coldStartThreshold = 5
	defaultRecLimit    = 20
	defaultSimLimit    = 10
	sectionItemLimit   = 10
	trendingWindow     = 7 * 24 * time.Hour
)

// Engine is the main recommendation engine that orchestrates all sub-engines.
// It implements the Recommender interface.
type Engine struct {
	repo         *repository.Repository
	cache        *cache.Cache
	contentBased *ContentBasedEngine
	collab       *CollaborativeEngine
	popularity   *PopularityEngine
}

// New creates a new Engine with all sub-engines initialized.
func New(repo *repository.Repository, cache *cache.Cache) *Engine {
	return &Engine{
		repo:         repo,
		cache:        cache,
		contentBased: NewContentBasedEngine(repo),
		collab:       NewCollaborativeEngine(repo),
		popularity:   NewPopularityEngine(repo),
	}
}

// ContentBased returns the content-based engine for batch computation.
func (e *Engine) ContentBased() *ContentBasedEngine { return e.contentBased }

// dynamicWeights returns (alpha, beta, gamma) weights based on interaction count.
func dynamicWeights(interactionCount int) (alpha, beta, gamma float64) {
	switch {
	case interactionCount < coldStartThreshold:
		return 0.0, 0.1, 0.9
	case interactionCount <= 20:
		return 0.2, 0.5, 0.3
	case interactionCount <= 50:
		return 0.4, 0.4, 0.2
	default:
		return 0.5, 0.35, 0.15
	}
}

// dominantAlgorithm returns the algorithm label based on which component contributed most.
func dominantAlgorithm(collabContrib, contentContrib, popContrib float64) string {
	max := collabContrib
	label := "collaborative"

	if contentContrib > max {
		max = contentContrib
		label = "content_based"
	}
	if popContrib > max {
		label = "popularity"
	}

	// If the top two are close (within 20% of each other), it's hybrid
	scores := []float64{collabContrib, contentContrib, popContrib}
	sort.Float64s(scores)
	if scores[2] > 0 && (scores[2]-scores[1])/scores[2] < 0.2 {
		return "hybrid"
	}
	return label
}

// GetRecommendations implements Recommender.
func (e *Engine) GetRecommendations(ctx context.Context, userID string, limit int) ([]RecommendedItem, error) {
	if limit <= 0 {
		limit = defaultRecLimit
	}

	// Check cache
	cached, err := e.cache.GetRecommendations(ctx, userID)
	if err == nil && cached != nil && len(cached.Items) > 0 {
		return e.enrichCachedRecommendations(ctx, cached, limit)
	}

	// Count interactions to decide strategy
	interactionCount, err := e.repo.CountUserInteractions(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("failed to count interactions, using popularity")
		interactionCount = 0
	}

	var items []RecommendedItem

	if interactionCount < coldStartThreshold {
		metrics.RecommendationColdStartFallbackTotal.Inc()
		items, err = e.popularityRecommend(ctx, userID, limit)
	} else {
		items, err = e.hybridRecommend(ctx, userID, interactionCount, limit)
	}

	if err != nil {
		return nil, err
	}

	// Cache results
	e.cacheRecommendations(ctx, userID, items)

	return items, nil
}

// popularityRecommend uses only the popularity engine.
func (e *Engine) popularityRecommend(ctx context.Context, userID string, limit int) ([]RecommendedItem, error) {
	scored, err := e.popularity.Recommend(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("popularity recommendations: %w", err)
	}

	contentMap, err := e.loadContentMap(ctx)
	if err != nil {
		return nil, err
	}

	var items []RecommendedItem
	for _, s := range scored {
		cf := contentMap[s.ContentID]
		if cf == nil {
			continue
		}
		items = append(items, contentFeaturesToRecommendedItem(cf, s.Score, "popularity", "Popular on StreamVault"))
	}
	return items, nil
}

// hybridRecommend combines all three sub-engines with dynamic weights.
func (e *Engine) hybridRecommend(ctx context.Context, userID string, interactionCount, limit int) ([]RecommendedItem, error) {
	alpha, beta, gamma := dynamicWeights(interactionCount)

	// Collect scores from each engine
	collabScores := make(map[string]float64)
	contentScores := make(map[string]float64)
	popScores := make(map[string]float64)

	// Collaborative filtering
	if alpha > 0 {
		collabItems, err := e.collab.Recommend(ctx, userID, limit*3)
		if err != nil {
			log.Warn().Err(err).Msg("collaborative filtering failed, skipping")
		} else {
			for _, item := range collabItems {
				collabScores[item.ContentID] = item.Score
			}
		}
	}

	// Content-based filtering
	if beta > 0 {
		contentItems, err := e.contentBased.Recommend(ctx, userID, limit*3)
		if err != nil {
			log.Warn().Err(err).Msg("content-based filtering failed, skipping")
		} else {
			for _, item := range contentItems {
				contentScores[item.ContentID] = item.Score
			}
		}
	}

	// Popularity
	if gamma > 0 {
		popItems, err := e.popularity.Recommend(ctx, userID, limit*3)
		if err != nil {
			log.Warn().Err(err).Msg("popularity engine failed, skipping")
		} else {
			for _, item := range popItems {
				popScores[item.ContentID] = item.Score
			}
		}
	}

	// Union of all candidate content IDs
	candidates := make(map[string]bool)
	for id := range collabScores {
		candidates[id] = true
	}
	for id := range contentScores {
		candidates[id] = true
	}
	for id := range popScores {
		candidates[id] = true
	}

	// Compute hybrid scores
	type hybridCandidate struct {
		ContentID      string
		FinalScore     float64
		CollabContrib  float64
		ContentContrib float64
		PopContrib     float64
	}

	var hybridCandidates []hybridCandidate
	for contentID := range candidates {
		cs := collabScores[contentID]
		bs := contentScores[contentID]
		ps := popScores[contentID]

		finalScore := alpha*cs + beta*bs + gamma*ps
		if finalScore > 0 {
			hybridCandidates = append(hybridCandidates, hybridCandidate{
				ContentID:      contentID,
				FinalScore:     finalScore,
				CollabContrib:  alpha * cs,
				ContentContrib: beta * bs,
				PopContrib:     gamma * ps,
			})
		}
	}

	sort.Slice(hybridCandidates, func(i, j int) bool {
		return hybridCandidates[i].FinalScore > hybridCandidates[j].FinalScore
	})

	if len(hybridCandidates) > limit {
		hybridCandidates = hybridCandidates[:limit]
	}

	// Load content metadata for enrichment
	contentMap, err := e.loadContentMap(ctx)
	if err != nil {
		return nil, err
	}

	// Build reason from user's watch history
	reason := e.buildReason(ctx, userID)

	var items []RecommendedItem
	for _, hc := range hybridCandidates {
		cf := contentMap[hc.ContentID]
		if cf == nil {
			continue
		}
		algorithm := dominantAlgorithm(hc.CollabContrib, hc.ContentContrib, hc.PopContrib)
		items = append(items, contentFeaturesToRecommendedItem(cf, hc.FinalScore, algorithm, reason))
	}
	return items, nil
}

// buildReason generates a reason string based on user's recent watch history.
func (e *Engine) buildReason(ctx context.Context, userID string) string {
	interactions, err := e.repo.GetUserInteractions(ctx, userID, 5)
	if err != nil || len(interactions) == 0 {
		return "Recommended for you"
	}

	// Find the most recent completed interaction with content features
	contentMap, _ := e.loadContentMap(ctx)
	for _, interaction := range interactions {
		if interaction.InteractionType == model.InteractionComplete ||
			interaction.CompletionPct >= 50 {
			if cf, ok := contentMap[interaction.ContentID]; ok {
				return fmt.Sprintf("Because you watched %s", cf.Title)
			}
		}
	}
	return "Recommended for you"
}

// GetSimilar implements Recommender.
func (e *Engine) GetSimilar(ctx context.Context, contentID string, limit int) ([]SimilarItem, error) {
	if limit <= 0 {
		limit = defaultSimLimit
	}

	// Check cache
	cached, err := e.cache.GetSimilarContent(ctx, contentID)
	if err == nil && cached != nil && len(cached.Items) > 0 {
		return e.enrichCachedSimilar(ctx, cached, limit)
	}

	// Get similar from content-based engine
	scored, err := e.contentBased.GetSimilar(ctx, contentID, limit, nil)
	if err != nil {
		return nil, fmt.Errorf("getting similar content: %w", err)
	}

	contentMap, err := e.loadContentMap(ctx)
	if err != nil {
		return nil, err
	}

	var items []SimilarItem
	for _, s := range scored {
		cf := contentMap[s.ContentID]
		if cf == nil {
			continue
		}
		items = append(items, SimilarItem{
			ContentID:       cf.ContentID,
			Title:           cf.Title,
			ContentType:     contentTypeFromGenres(cf.Genres),
			ReleaseYear:     cf.ReleaseYear,
			AverageRating:   cf.AvgRating,
			SimilarityScore: s.Score,
		})
	}

	// Cache results
	e.cacheSimilar(ctx, contentID, items)

	return items, nil
}

// GetHomePageSections implements Recommender.
func (e *Engine) GetHomePageSections(ctx context.Context, userID string) ([]HomePageSection, error) {
	var sections []HomePageSection

	// PERSONAL section (skip for anonymous/cold-start users)
	if userID != "" {
		count, _ := e.repo.CountUserInteractions(ctx, userID)
		if count >= coldStartThreshold {
			recs, err := e.GetRecommendations(ctx, userID, sectionItemLimit)
			if err == nil && len(recs) > 0 {
				sections = append(sections, HomePageSection{
					SectionType: "personal",
					Title:       "Recommended For You",
					Items:       recs,
				})
			}
		}
	}

	// TRENDING section
	trendingSection, err := e.buildTrendingSection(ctx, sectionItemLimit)
	if err == nil && len(trendingSection) > 0 {
		sections = append(sections, HomePageSection{
			SectionType: "trending",
			Title:       "Trending Now",
			Items:       trendingSection,
		})
	}

	// BECAUSE_YOU_WATCHED section
	if userID != "" {
		bywSection, err := e.buildBecauseYouWatchedSection(ctx, userID, sectionItemLimit)
		if err == nil && bywSection != nil {
			sections = append(sections, *bywSection)
		}
	}

	// GENRE section
	if userID != "" {
		genreSection, err := e.buildGenreSection(ctx, userID, sectionItemLimit)
		if err == nil && genreSection != nil {
			sections = append(sections, *genreSection)
		}
	}

	// NEW section
	newSection, err := e.buildNewReleasesSection(ctx, sectionItemLimit)
	if err == nil && len(newSection) > 0 {
		sections = append(sections, HomePageSection{
			SectionType: "new",
			Title:       "New Releases",
			Items:       newSection,
		})
	}

	return sections, nil
}

// RecordFeedback implements Recommender.
func (e *Engine) RecordFeedback(ctx context.Context, userID, contentID, feedbackType string) error {
	interaction := &model.Interaction{
		UserID:          userID,
		ContentID:       contentID,
		InteractionType: model.InteractionType(feedbackType),
		CreatedAt:       time.Now(),
	}
	if err := e.repo.CreateInteraction(ctx, interaction); err != nil {
		return fmt.Errorf("recording feedback: %w", err)
	}
	e.cache.InvalidateRecommendations(ctx, userID)
	return nil
}

// --- Home page section builders ---

func (e *Engine) buildTrendingSection(ctx context.Context, limit int) ([]RecommendedItem, error) {
	since := time.Now().Add(-trendingWindow)
	viewCounts, err := e.repo.GetRecentViewCounts(ctx, since)
	if err != nil {
		return nil, err
	}

	contentMap, err := e.loadContentMap(ctx)
	if err != nil {
		return nil, err
	}

	type viewItem struct {
		contentID string
		views     int
	}
	var items []viewItem
	for id, cnt := range viewCounts {
		items = append(items, viewItem{id, cnt})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].views > items[j].views
	})

	if len(items) == 0 {
		// Fallback: return popular content by rating if no recent views
		return e.buildNewReleasesSection(ctx, limit)
	}

	if len(items) > limit {
		items = items[:limit]
	}

	maxViews := items[0].views
	var result []RecommendedItem
	for _, vi := range items {
		cf := contentMap[vi.contentID]
		if cf == nil {
			continue
		}
		score := float64(vi.views) / float64(maxViews)
		result = append(result, contentFeaturesToRecommendedItem(cf, score, "popularity", "Trending Now"))
	}
	return result, nil
}

func (e *Engine) buildBecauseYouWatchedSection(ctx context.Context, userID string, limit int) (*HomePageSection, error) {
	interactions, err := e.repo.GetUserInteractions(ctx, userID, 20)
	if err != nil || len(interactions) == 0 {
		return nil, err
	}

	contentMap, err := e.loadContentMap(ctx)
	if err != nil {
		return nil, err
	}

	// Find most recent fully watched content
	var watchedTitle string
	var watchedContentID string
	for _, i := range interactions {
		if i.InteractionType == model.InteractionComplete || i.CompletionPct >= 90 {
			if cf, ok := contentMap[i.ContentID]; ok {
				watchedContentID = i.ContentID
				watchedTitle = cf.Title
				break
			}
		}
	}

	if watchedContentID == "" {
		return nil, nil
	}

	similar, err := e.GetSimilar(ctx, watchedContentID, limit)
	if err != nil || len(similar) == 0 {
		return nil, err
	}

	// Convert SimilarItem to RecommendedItem
	var items []RecommendedItem
	for _, s := range similar {
		items = append(items, RecommendedItem{
			ContentID:     s.ContentID,
			Title:         s.Title,
			ContentType:   s.ContentType,
			ReleaseYear:   s.ReleaseYear,
			AverageRating: s.AverageRating,
			Score:         s.SimilarityScore,
			Algorithm:     "content_based",
			Reason:        fmt.Sprintf("Because you watched %s", watchedTitle),
		})
	}

	return &HomePageSection{
		SectionType: "because_you_watched",
		Title:       fmt.Sprintf("Because You Watched %s", watchedTitle),
		Items:       items,
	}, nil
}

func (e *Engine) buildGenreSection(ctx context.Context, userID string, limit int) (*HomePageSection, error) {
	profile, err := e.repo.GetUserProfile(ctx, userID)
	if err != nil || profile == nil || len(profile.GenreWeights) == 0 {
		return nil, err
	}

	// Find top genre
	var topGenre string
	var topWeight float64
	for genre, weight := range profile.GenreWeights {
		if weight > topWeight {
			topWeight = weight
			topGenre = genre
		}
	}

	if topGenre == "" {
		return nil, nil
	}

	allContent, err := e.repo.GetAllContentFeatures(ctx)
	if err != nil {
		return nil, err
	}

	// Filter to top genre, sort by rating
	var genreContent []model.ContentFeatures
	for _, cf := range allContent {
		for _, g := range cf.Genres {
			if g == topGenre {
				genreContent = append(genreContent, cf)
				break
			}
		}
	}

	sort.Slice(genreContent, func(i, j int) bool {
		return genreContent[i].AvgRating > genreContent[j].AvgRating
	})

	if len(genreContent) > limit {
		genreContent = genreContent[:limit]
	}

	var items []RecommendedItem
	for _, cf := range genreContent {
		score := cf.AvgRating / 10.0
		items = append(items, contentFeaturesToRecommendedItem(&cf, score, "content_based", fmt.Sprintf("Top %s picks", topGenre)))
	}

	if len(items) == 0 {
		return nil, nil
	}

	return &HomePageSection{
		SectionType: "genre",
		Title:       fmt.Sprintf("Top %s Picks", topGenre),
		Items:       items,
	}, nil
}

func (e *Engine) buildNewReleasesSection(ctx context.Context, limit int) ([]RecommendedItem, error) {
	allContent, err := e.repo.GetAllContentFeatures(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(allContent, func(i, j int) bool {
		if allContent[i].ReleaseYear == allContent[j].ReleaseYear {
			return allContent[i].AvgRating > allContent[j].AvgRating
		}
		return allContent[i].ReleaseYear > allContent[j].ReleaseYear
	})

	if len(allContent) > limit {
		allContent = allContent[:limit]
	}

	var items []RecommendedItem
	for _, cf := range allContent {
		score := cf.AvgRating / 10.0
		items = append(items, contentFeaturesToRecommendedItem(&cf, score, "popularity", "New Release"))
	}
	return items, nil
}

// --- Helper functions ---

func (e *Engine) loadContentMap(ctx context.Context) (map[string]*model.ContentFeatures, error) {
	allContent, err := e.repo.GetAllContentFeatures(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading content features: %w", err)
	}

	contentMap := make(map[string]*model.ContentFeatures, len(allContent))
	for idx := range allContent {
		contentMap[allContent[idx].ContentID] = &allContent[idx]
	}
	return contentMap, nil
}

func contentFeaturesToRecommendedItem(cf *model.ContentFeatures, score float64, algorithm, reason string) RecommendedItem {
	return RecommendedItem{
		ContentID:     cf.ContentID,
		Title:         cf.Title,
		ContentType:   contentTypeFromGenres(cf.Genres),
		ReleaseYear:   cf.ReleaseYear,
		AverageRating: cf.AvgRating,
		Score:         score,
		Algorithm:     algorithm,
		Reason:        reason,
	}
}

// contentTypeFromGenres infers the content type. Since content_features doesn't
// store content type directly, we default to "movie".
func contentTypeFromGenres(_ []string) string {
	return "movie"
}

func (e *Engine) enrichCachedRecommendations(ctx context.Context, cached *cache.RecommendationResult, limit int) ([]RecommendedItem, error) {
	contentMap, err := e.loadContentMap(ctx)
	if err != nil {
		return nil, err
	}

	var items []RecommendedItem
	for _, ci := range cached.Items {
		cf := contentMap[ci.ContentID]
		if cf == nil {
			continue
		}
		items = append(items, contentFeaturesToRecommendedItem(cf, ci.Score, ci.Algorithm, ci.Reason))
	}

	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (e *Engine) enrichCachedSimilar(ctx context.Context, cached *cache.SimilarResult, limit int) ([]SimilarItem, error) {
	contentMap, err := e.loadContentMap(ctx)
	if err != nil {
		return nil, err
	}

	var items []SimilarItem
	for _, ci := range cached.Items {
		cf := contentMap[ci.ContentID]
		if cf == nil {
			continue
		}
		items = append(items, SimilarItem{
			ContentID:       cf.ContentID,
			Title:           cf.Title,
			ContentType:     contentTypeFromGenres(cf.Genres),
			ReleaseYear:     cf.ReleaseYear,
			AverageRating:   cf.AvgRating,
			SimilarityScore: ci.SimilarityScore,
		})
	}

	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (e *Engine) cacheRecommendations(ctx context.Context, userID string, items []RecommendedItem) {
	cacheItems := make([]cache.RecommendedItem, len(items))
	for i, item := range items {
		cacheItems[i] = cache.RecommendedItem{
			ContentID: item.ContentID,
			Score:     item.Score,
			Algorithm: item.Algorithm,
			Reason:    item.Reason,
		}
	}
	e.cache.SetRecommendations(ctx, userID, &cache.RecommendationResult{Items: cacheItems})
}

func (e *Engine) cacheSimilar(ctx context.Context, contentID string, items []SimilarItem) {
	cacheItems := make([]cache.SimilarItem, len(items))
	for i, item := range items {
		cacheItems[i] = cache.SimilarItem{
			ContentID:       item.ContentID,
			SimilarityScore: item.SimilarityScore,
		}
	}
	e.cache.SetSimilarContent(ctx, contentID, &cache.SimilarResult{Items: cacheItems})
}
