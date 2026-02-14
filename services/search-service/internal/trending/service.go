package trending

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/model"
)

// Service manages trending content data backed by Redis sorted sets.
type Service struct {
	client *redis.Client
}

// NewService creates a new trending Service.
func NewService(client *redis.Client) *Service {
	return &Service{client: client}
}

// GetTrending returns the top trending items for a given time window.
func (s *Service) GetTrending(ctx context.Context, window string, limit int) (*model.TrendingResponse, error) {
	key := windowKey(window)
	prevKey := prevWindowKey(window)

	results, err := s.client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("fetching trending from redis: %w", err)
	}

	items := make([]model.TrendingItem, 0, len(results))
	for i, z := range results {
		contentID, ok := z.Member.(string)
		if !ok {
			continue
		}

		item := model.TrendingItem{
			ContentID: contentID,
			Rank:      i + 1,
			ViewCount: int64(z.Score),
		}

		meta, err := s.client.HGetAll(ctx, fmt.Sprintf("trending:meta:%s", contentID)).Result()
		if err == nil && len(meta) > 0 {
			item.Title = meta["title"]
			item.ThumbnailURL = meta["thumbnail_url"]
			item.ContentType = meta["content_type"]
			item.ReleaseYear, _ = strconv.Atoi(meta["release_year"])
			item.AverageRating, _ = strconv.ParseFloat(meta["average_rating"], 64)
		}

		prevRank, err := s.client.ZRevRank(ctx, prevKey, contentID).Result()
		if err == nil {
			item.RankChange = int(prevRank+1) - item.Rank
		}

		items = append(items, item)
	}

	return &model.TrendingResponse{Items: items}, nil
}

// SetContentMeta caches content metadata used for trending display.
func (s *Service) SetContentMeta(ctx context.Context, contentID string, title, thumbnailURL, contentType string, releaseYear int, averageRating float64) {
	key := fmt.Sprintf("trending:meta:%s", contentID)
	fields := map[string]interface{}{
		"title":          title,
		"thumbnail_url":  thumbnailURL,
		"content_type":   contentType,
		"release_year":   releaseYear,
		"average_rating": averageRating,
	}
	if err := s.client.HSet(ctx, key, fields).Err(); err != nil {
		log.Warn().Err(err).Str("content_id", contentID).Msg("failed to set trending metadata")
	}
}

// IncrementViewCount increments trending scores for all time windows.
func (s *Service) IncrementViewCount(ctx context.Context, contentID string) {
	pipe := s.client.Pipeline()
	pipe.ZIncrBy(ctx, "trending:daily", 1, contentID)
	pipe.ZIncrBy(ctx, "trending:weekly", 1, contentID)
	pipe.ZIncrBy(ctx, "trending:monthly", 1, contentID)
	if _, err := pipe.Exec(ctx); err != nil {
		log.Warn().Err(err).Str("content_id", contentID).Msg("failed to increment trending scores")
	}
}

func windowKey(window string) string {
	switch window {
	case "day":
		return "trending:daily"
	case "month":
		return "trending:monthly"
	default:
		return "trending:weekly"
	}
}

func prevWindowKey(window string) string {
	switch window {
	case "day":
		return "trending:daily:prev"
	case "month":
		return "trending:monthly:prev"
	default:
		return "trending:weekly:prev"
	}
}
