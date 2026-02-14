package progress

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

const completionThreshold = 95.0
const watchCompletedThreshold = 90.0

type WatchCompletedPublisher interface {
	PublishWatchCompleted(ctx context.Context, userID, contentID string, percentage float64) error
}

type Service struct {
	repo      Repository
	publisher WatchCompletedPublisher
}

func NewService(repo Repository, publisher WatchCompletedPublisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

func (s *Service) SaveProgress(ctx context.Context, userID, contentID string, positionSeconds, durationSeconds int64) error {
	if durationSeconds <= 0 {
		durationSeconds = 1
	}

	percentage := float64(positionSeconds) / float64(durationSeconds) * 100

	// Publish watch completed event when >= 90%
	if percentage >= watchCompletedThreshold && s.publisher != nil {
		if err := s.publisher.PublishWatchCompleted(ctx, userID, contentID, percentage); err != nil {
			log.Warn().Err(err).
				Str("userId", userID).
				Str("contentId", contentID).
				Msg("failed to publish watch completed event")
		}
	}

	// If completed (>= 95%), remove from continue-watching
	if percentage >= completionThreshold {
		return s.repo.Remove(ctx, userID, contentID)
	}

	p := WatchProgress{
		UserID:          userID,
		ContentID:       contentID,
		PositionSeconds: positionSeconds,
		DurationSeconds: durationSeconds,
		Percentage:      percentage,
		UpdatedAt:       time.Now().Unix(),
	}

	return s.repo.Save(ctx, p)
}

func (s *Service) GetProgress(ctx context.Context, userID, contentID string) (*WatchProgress, error) {
	return s.repo.Get(ctx, userID, contentID)
}

func (s *Service) GetContinueWatching(ctx context.Context, userID string, limit int) ([]WatchProgress, error) {
	return s.repo.GetContinueWatching(ctx, userID, limit)
}
