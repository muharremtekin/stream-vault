package handler

import (
	"context"
	"io"
	"time"

	"github.com/streamvault/streaming-service/internal/messaging"
	"github.com/streamvault/streaming-service/internal/progress"
	"github.com/streamvault/streaming-service/internal/storage"
)

// mockProgressRepo implements progress.Repository for testing.
type mockProgressRepo struct {
	saveFunc              func(ctx context.Context, p progress.WatchProgress) error
	getFunc               func(ctx context.Context, userID, contentID string) (*progress.WatchProgress, error)
	getContinueWatching   func(ctx context.Context, userID string, limit int) ([]progress.WatchProgress, error)
	removeFunc            func(ctx context.Context, userID, contentID string) error
}

func (m *mockProgressRepo) Save(ctx context.Context, p progress.WatchProgress) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, p)
	}
	return nil
}

func (m *mockProgressRepo) Get(ctx context.Context, userID, contentID string) (*progress.WatchProgress, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, userID, contentID)
	}
	return nil, nil
}

func (m *mockProgressRepo) GetContinueWatching(ctx context.Context, userID string, limit int) ([]progress.WatchProgress, error) {
	if m.getContinueWatching != nil {
		return m.getContinueWatching(ctx, userID, limit)
	}
	return nil, nil
}

func (m *mockProgressRepo) Remove(ctx context.Context, userID, contentID string) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, userID, contentID)
	}
	return nil
}

// mockStorage implements storage.Storage for testing.
type mockStorage struct {
	uploadFunc   func(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
	downloadFunc func(ctx context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error)
	listFunc     func(ctx context.Context, bucket, prefix string) ([]storage.ObjectInfo, error)
}

func (m *mockStorage) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, bucket, key, reader, size, contentType)
	}
	return nil
}

func (m *mockStorage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, bucket, key)
	}
	return io.NopCloser(io.LimitReader(nil, 0)), &storage.ObjectInfo{Size: 0}, nil
}

func (m *mockStorage) List(ctx context.Context, bucket, prefix string) ([]storage.ObjectInfo, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, bucket, prefix)
	}
	return nil, nil
}

func (m *mockStorage) Delete(ctx context.Context, bucket, key string) error { return nil }
func (m *mockStorage) Exists(ctx context.Context, bucket, key string) (bool, error) {
	return false, nil
}
func (m *mockStorage) HealthCheck(ctx context.Context) error { return nil }

// mockPublisher implements messaging.Publisher for testing.
type mockPublisher struct {
	publishFunc func(ctx context.Context, job messaging.EncodingJob) error
	lastJob     *messaging.EncodingJob
}

func (m *mockPublisher) PublishEncodingJob(ctx context.Context, job messaging.EncodingJob) error {
	m.lastJob = &job
	if m.publishFunc != nil {
		return m.publishFunc(ctx, job)
	}
	return nil
}

func (m *mockPublisher) PublishWatchCompleted(ctx context.Context, event messaging.WatchCompletedEvent) error {
	return nil
}

func (m *mockPublisher) Close() error { return nil }

// Ensure compile-time interface satisfaction.
var _ storage.Storage = (*mockStorage)(nil)
var _ messaging.Publisher = (*mockPublisher)(nil)

// suppress unused import warnings
var _ = time.Now
