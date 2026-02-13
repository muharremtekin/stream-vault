package progress

import (
	"context"
	"time"
)

const completionThreshold = 95.0

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SaveProgress(ctx context.Context, userID, contentID string, positionSeconds, durationSeconds int64) error {
	if durationSeconds <= 0 {
		durationSeconds = 1
	}

	percentage := float64(positionSeconds) / float64(durationSeconds) * 100

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
