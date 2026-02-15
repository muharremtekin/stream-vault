package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/metrics"
	"github.com/streamvault/recommendation-service/internal/model"
)

const (
	userProfileTTL     = 1 * time.Hour
	contentFeaturesTTL = 6 * time.Hour
	recommendationsTTL = 30 * time.Minute
	similarContentTTL  = 6 * time.Hour
)

// Cache provides Redis-based caching for the recommendation engine.
type Cache struct {
	client *redis.Client
}

// New creates a new Cache.
func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// --- User Profile Cache ---

// GetUserProfile retrieves a cached user profile.
func (c *Cache) GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	key := fmt.Sprintf("rec:user_profile:%s", userID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		metrics.RecommendationCacheMissesTotal.Inc()
		return nil, err
	}
	var profile model.UserProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("unmarshalling cached user profile: %w", err)
	}
	metrics.RecommendationCacheHitsTotal.Inc()
	return &profile, nil
}

// SetUserProfile stores a user profile in cache.
func (c *Cache) SetUserProfile(ctx context.Context, profile *model.UserProfile) {
	key := fmt.Sprintf("rec:user_profile:%s", profile.UserID)
	data, err := json.Marshal(profile)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal user profile for cache")
		return
	}
	if err := c.client.Set(ctx, key, data, userProfileTTL).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", profile.UserID).Msg("failed to cache user profile")
	}
}

// InvalidateUserProfile removes a user profile from cache.
func (c *Cache) InvalidateUserProfile(ctx context.Context, userID string) {
	key := fmt.Sprintf("rec:user_profile:%s", userID)
	c.client.Del(ctx, key)
}

// --- Content Features Cache ---

// GetContentFeatures retrieves cached content features.
func (c *Cache) GetContentFeatures(ctx context.Context, contentID string) (*model.ContentFeatures, error) {
	key := fmt.Sprintf("rec:content_features:%s", contentID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		metrics.RecommendationCacheMissesTotal.Inc()
		return nil, err
	}
	var features model.ContentFeatures
	if err := json.Unmarshal(data, &features); err != nil {
		return nil, fmt.Errorf("unmarshalling cached content features: %w", err)
	}
	metrics.RecommendationCacheHitsTotal.Inc()
	return &features, nil
}

// SetContentFeatures stores content features in cache.
func (c *Cache) SetContentFeatures(ctx context.Context, features *model.ContentFeatures) {
	key := fmt.Sprintf("rec:content_features:%s", features.ContentID)
	data, err := json.Marshal(features)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal content features for cache")
		return
	}
	if err := c.client.Set(ctx, key, data, contentFeaturesTTL).Err(); err != nil {
		log.Warn().Err(err).Str("content_id", features.ContentID).Msg("failed to cache content features")
	}
}

// InvalidateContentFeatures removes content features from cache.
func (c *Cache) InvalidateContentFeatures(ctx context.Context, contentID string) {
	key := fmt.Sprintf("rec:content_features:%s", contentID)
	c.client.Del(ctx, key)
}

// --- Recommendation Results Cache ---

// RecommendationResult wraps a list of recommended items for caching.
type RecommendationResult struct {
	Items []RecommendedItem `json:"items"`
}

// RecommendedItem is a simplified recommendation for caching.
type RecommendedItem struct {
	ContentID string  `json:"contentId"`
	Score     float64 `json:"score"`
	Algorithm string  `json:"algorithm"`
	Reason    string  `json:"reason"`
}

// GetRecommendations retrieves cached recommendations for a user.
func (c *Cache) GetRecommendations(ctx context.Context, userID string) (*RecommendationResult, error) {
	key := fmt.Sprintf("rec:recommendations:%s", userID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		metrics.RecommendationCacheMissesTotal.Inc()
		return nil, err
	}
	var result RecommendationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshalling cached recommendations: %w", err)
	}
	metrics.RecommendationCacheHitsTotal.Inc()
	return &result, nil
}

// SetRecommendations stores recommendations in cache.
func (c *Cache) SetRecommendations(ctx context.Context, userID string, result *RecommendationResult) {
	key := fmt.Sprintf("rec:recommendations:%s", userID)
	data, err := json.Marshal(result)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal recommendations for cache")
		return
	}
	if err := c.client.Set(ctx, key, data, recommendationsTTL).Err(); err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("failed to cache recommendations")
	}
}

// InvalidateRecommendations removes cached recommendations for a user.
func (c *Cache) InvalidateRecommendations(ctx context.Context, userID string) {
	key := fmt.Sprintf("rec:recommendations:%s", userID)
	c.client.Del(ctx, key)
}

// --- Similar Content Cache ---

// SimilarResult wraps a list of similar items for caching.
type SimilarResult struct {
	Items []SimilarItem `json:"items"`
}

// SimilarItem is a simplified similar content entry for caching.
type SimilarItem struct {
	ContentID       string  `json:"contentId"`
	SimilarityScore float64 `json:"similarityScore"`
}

// GetSimilarContent retrieves cached similar content.
func (c *Cache) GetSimilarContent(ctx context.Context, contentID string) (*SimilarResult, error) {
	key := fmt.Sprintf("rec:similar:%s", contentID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		metrics.RecommendationCacheMissesTotal.Inc()
		return nil, err
	}
	var result SimilarResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshalling cached similar content: %w", err)
	}
	metrics.RecommendationCacheHitsTotal.Inc()
	return &result, nil
}

// SetSimilarContent stores similar content in cache.
func (c *Cache) SetSimilarContent(ctx context.Context, contentID string, result *SimilarResult) {
	key := fmt.Sprintf("rec:similar:%s", contentID)
	data, err := json.Marshal(result)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal similar content for cache")
		return
	}
	if err := c.client.Set(ctx, key, data, similarContentTTL).Err(); err != nil {
		log.Warn().Err(err).Str("content_id", contentID).Msg("failed to cache similar content")
	}
}

// InvalidateSimilarContent removes cached similar content.
func (c *Cache) InvalidateSimilarContent(ctx context.Context, contentID string) {
	key := fmt.Sprintf("rec:similar:%s", contentID)
	c.client.Del(ctx, key)
}
