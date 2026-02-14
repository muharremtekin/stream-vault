package progress

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/streamvault/streaming-service/internal/messaging"
)

type watchCompletedPublisher struct {
	publisher messaging.Publisher
	redis     *redis.Client
	dedupTTL  time.Duration
}

func NewWatchCompletedPublisher(pub messaging.Publisher, redisClient *redis.Client, dedupTTL time.Duration) WatchCompletedPublisher {
	return &watchCompletedPublisher{
		publisher: pub,
		redis:     redisClient,
		dedupTTL:  dedupTTL,
	}
}

func (w *watchCompletedPublisher) PublishWatchCompleted(ctx context.Context, userID, contentID string, percentage float64) error {
	dedupKey := fmt.Sprintf("watch-completed:%s:%s", userID, contentID)

	ok, err := w.redis.SetNX(ctx, dedupKey, "1", w.dedupTTL).Result()
	if err != nil {
		return fmt.Errorf("checking dedup key: %w", err)
	}
	if !ok {
		return nil
	}

	now := time.Now().UTC()
	event := messaging.WatchCompletedEvent{
		EventID:       uuid.New().String(),
		EventType:     "watch.completed",
		Timestamp:     now.Format(time.RFC3339),
		Source:        "streaming-service",
		CorrelationID: uuid.New().String(),
		Data: messaging.WatchCompletedData{
			UserID:               userID,
			ContentID:            contentID,
			CompletionPercentage: percentage,
			WatchedAt:            now.Format(time.RFC3339),
		},
	}

	return w.publisher.PublishWatchCompleted(ctx, event)
}
