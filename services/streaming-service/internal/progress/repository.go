package progress

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type WatchProgress struct {
	UserID          string  `json:"user_id"`
	ContentID       string  `json:"content_id"`
	PositionSeconds int64   `json:"position_seconds"`
	DurationSeconds int64   `json:"duration_seconds"`
	Percentage      float64 `json:"percentage"`
	UpdatedAt       int64   `json:"updated_at"`
}

type Repository interface {
	Save(ctx context.Context, p WatchProgress) error
	Get(ctx context.Context, userID, contentID string) (*WatchProgress, error)
	GetContinueWatching(ctx context.Context, userID string, limit int) ([]WatchProgress, error)
	Remove(ctx context.Context, userID, contentID string) error
}

type redisRepository struct {
	client      *redis.Client
	progressTTL time.Duration
}

func NewRedisRepository(client *redis.Client, progressTTL time.Duration) Repository {
	return &redisRepository{
		client:      client,
		progressTTL: progressTTL,
	}
}

func progressKey(userID, contentID string) string {
	return fmt.Sprintf("progress:%s:%s", userID, contentID)
}

func continueWatchingKey(userID string) string {
	return fmt.Sprintf("continue-watching:%s", userID)
}

func (r *redisRepository) Save(ctx context.Context, p WatchProgress) error {
	key := progressKey(p.UserID, p.ContentID)

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"position_seconds": p.PositionSeconds,
		"duration_seconds": p.DurationSeconds,
		"percentage":       p.Percentage,
		"updated_at":       p.UpdatedAt,
	})
	pipe.Expire(ctx, key, r.progressTTL)

	// Update continue-watching sorted set
	cwKey := continueWatchingKey(p.UserID)
	pipe.ZAdd(ctx, cwKey, redis.Z{
		Score:  float64(p.UpdatedAt),
		Member: p.ContentID,
	})

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("saving progress: %w", err)
	}

	return nil
}

func (r *redisRepository) Get(ctx context.Context, userID, contentID string) (*WatchProgress, error) {
	key := progressKey(userID, contentID)

	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("getting progress: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}

	positionSeconds, _ := strconv.ParseInt(result["position_seconds"], 10, 64)
	durationSeconds, _ := strconv.ParseInt(result["duration_seconds"], 10, 64)
	percentage, _ := strconv.ParseFloat(result["percentage"], 64)
	updatedAt, _ := strconv.ParseInt(result["updated_at"], 10, 64)

	return &WatchProgress{
		UserID:          userID,
		ContentID:       contentID,
		PositionSeconds: positionSeconds,
		DurationSeconds: durationSeconds,
		Percentage:      percentage,
		UpdatedAt:       updatedAt,
	}, nil
}

func (r *redisRepository) GetContinueWatching(ctx context.Context, userID string, limit int) ([]WatchProgress, error) {
	if limit <= 0 {
		limit = 20
	}

	cwKey := continueWatchingKey(userID)

	// Get content IDs sorted by most recent
	contentIDs, err := r.client.ZRevRangeByScore(ctx, cwKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Count:  int64(limit),
		Offset: 0,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("getting continue watching list: %w", err)
	}

	var items []WatchProgress
	for _, contentID := range contentIDs {
		p, err := r.Get(ctx, userID, contentID)
		if err != nil {
			continue
		}
		if p == nil {
			// Progress expired, remove from sorted set
			r.client.ZRem(ctx, cwKey, contentID)
			continue
		}
		items = append(items, *p)
	}

	return items, nil
}

func (r *redisRepository) Remove(ctx context.Context, userID, contentID string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, progressKey(userID, contentID))
	pipe.ZRem(ctx, continueWatchingKey(userID), contentID)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("removing progress: %w", err)
	}
	return nil
}
