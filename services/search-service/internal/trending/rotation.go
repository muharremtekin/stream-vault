package trending

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// StartRotation runs a background ticker that rotates trending windows
// at daily, weekly (Monday), and monthly (1st) boundaries.
func (s *Service) StartRotation(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Info().Msg("trending rotation started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("trending rotation stopped")
			return
		case <-ticker.C:
			s.checkAndRotate(ctx)
		}
	}
}

func (s *Service) checkAndRotate(ctx context.Context) {
	now := time.Now().UTC()

	s.rotateIfNeeded(ctx, "daily", now.Format("2006-01-02"), "trending:daily", "trending:daily:prev")

	year, week := now.ISOWeek()
	weekKey := fmt.Sprintf("%d-W%02d", year, week)
	s.rotateIfNeeded(ctx, "weekly", weekKey, "trending:weekly", "trending:weekly:prev")

	monthKey := now.Format("2006-01")
	s.rotateIfNeeded(ctx, "monthly", monthKey, "trending:monthly", "trending:monthly:prev")
}

func (s *Service) rotateIfNeeded(ctx context.Context, window, currentPeriod, srcKey, destKey string) {
	lastKey := fmt.Sprintf("trending:last_rotation:%s", window)

	lastPeriod, err := s.client.Get(ctx, lastKey).Result()
	if err == nil && lastPeriod == currentPeriod {
		return
	}

	exists, _ := s.client.Exists(ctx, srcKey).Result()
	if exists > 0 {
		if err := s.client.Rename(ctx, srcKey, destKey).Err(); err != nil {
			log.Warn().Err(err).Str("window", window).Msg("failed to rotate trending window")
			return
		}
		log.Info().Str("window", window).Str("period", currentPeriod).Msg("trending window rotated")
	}

	s.client.Set(ctx, lastKey, currentPeriod, 0)

	// Trim previous window to top 200 to avoid unbounded growth
	s.client.ZRemRangeByRank(ctx, destKey, 0, -201)

	// Keep daily view count keys for 48h via TTL on individual keys (set by watch consumer)
	// Trim monthly sorted set entries older than 100 days via score if needed
	if window == "daily" {
		s.cleanupDailyViewCounts(ctx)
	}
}

func (s *Service) cleanupDailyViewCounts(ctx context.Context) {
	yesterday := time.Now().UTC().Add(-48 * time.Hour).Format("2006-01-02")
	pattern := fmt.Sprintf("views:daily:%s:*", yesterday)

	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()
	count := 0
	for iter.Next(ctx) {
		s.client.Del(ctx, iter.Val())
		count++
	}
	if count > 0 {
		log.Debug().Int("count", count).Str("date", yesterday).Msg("cleaned up old daily view counts")
	}
}
