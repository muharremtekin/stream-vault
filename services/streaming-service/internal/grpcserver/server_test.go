package grpcserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/streamvault/streaming-service/proto/common/v1"
	streamingv1 "github.com/streamvault/streaming-service/proto/streaming/v1"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/messaging"
	"github.com/streamvault/streaming-service/internal/progress"
	"github.com/streamvault/streaming-service/internal/storage"
)

// mockStorage implements storage.Storage for tests.
type mockStorage struct {
	listFunc     func(ctx context.Context, bucket, prefix string) ([]storage.ObjectInfo, error)
	downloadFunc func(ctx context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error)
}

func (m *mockStorage) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	return nil
}
func (m *mockStorage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *storage.ObjectInfo, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, bucket, key)
	}
	return nil, nil, errors.New("not implemented")
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

func setupServer(t *testing.T) (*StreamingServer, *redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)
	store := &mockStorage{}
	minioCfg := config.MinIOConfig{
		EncodedBucket: "streamvault-encoded",
	}
	server := NewStreamingServer(store, svc, client, minioCfg)
	return server, client, mr
}

// --- GetStreamingInfo tests ---

func TestGetStreamingInfo_EmptyContentId_ReturnsInvalidArgument(t *testing.T) {
	server, _, _ := setupServer(t)

	_, err := server.GetStreamingInfo(context.Background(), &streamingv1.GetStreamingInfoRequest{
		ContentId: "",
	})

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestGetStreamingInfo_CacheHit_ReturnsReadyWithQualities(t *testing.T) {
	server, client, _ := setupServer(t)
	ctx := context.Background()

	// Populate Redis cache
	result := messaging.EncodingResult{
		Duration: 120,
		Outputs: []messaging.EncodingOutput{
			{Quality: "720p", Width: 1280, Height: 720, BitrateKbps: 2800, SegmentCount: 12},
			{Quality: "1080p", Width: 1920, Height: 1080, BitrateKbps: 5000, SegmentCount: 12},
		},
	}
	data, _ := json.Marshal(result)
	client.Set(ctx, "stream-info:movie-1", string(data), 0)

	resp, err := server.GetStreamingInfo(ctx, &streamingv1.GetStreamingInfoRequest{
		ContentId: "movie-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != commonv1.VideoStatus_VIDEO_STATUS_READY {
		t.Errorf("expected READY status, got %v", resp.Status)
	}
	if resp.DurationSeconds != 120 {
		t.Errorf("expected duration 120, got %d", resp.DurationSeconds)
	}
	if len(resp.AvailableQualities) != 2 {
		t.Fatalf("expected 2 qualities, got %d", len(resp.AvailableQualities))
	}
	if resp.AvailableQualities[0].Label != "720p" {
		t.Errorf("expected first quality 720p, got %s", resp.AvailableQualities[0].Label)
	}
	if resp.ManifestPath != "/stream/movie-1/manifest.m3u8" {
		t.Errorf("unexpected manifest path: %s", resp.ManifestPath)
	}
}

func TestGetStreamingInfo_CacheMiss_ObjectsExist_ReturnsReady(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)

	store := &mockStorage{
		listFunc: func(ctx context.Context, bucket, prefix string) ([]storage.ObjectInfo, error) {
			return []storage.ObjectInfo{
				{Key: "movie-1/720p/playlist.m3u8"},
			}, nil
		},
	}

	server := NewStreamingServer(store, svc, client, config.MinIOConfig{
		EncodedBucket: "streamvault-encoded",
	})

	resp, err := server.GetStreamingInfo(context.Background(), &streamingv1.GetStreamingInfoRequest{
		ContentId: "movie-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != commonv1.VideoStatus_VIDEO_STATUS_READY {
		t.Errorf("expected READY status, got %v", resp.Status)
	}
	if resp.ManifestPath != "/stream/movie-1/manifest.m3u8" {
		t.Errorf("unexpected manifest path: %s", resp.ManifestPath)
	}
}

func TestGetStreamingInfo_CacheMiss_NoObjects_ReturnsNotUploaded(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)

	store := &mockStorage{
		listFunc: func(ctx context.Context, bucket, prefix string) ([]storage.ObjectInfo, error) {
			return nil, nil
		},
	}

	server := NewStreamingServer(store, svc, client, config.MinIOConfig{
		EncodedBucket: "streamvault-encoded",
	})

	resp, err := server.GetStreamingInfo(context.Background(), &streamingv1.GetStreamingInfoRequest{
		ContentId: "movie-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != commonv1.VideoStatus_VIDEO_STATUS_NOT_UPLOADED {
		t.Errorf("expected NOT_UPLOADED status, got %v", resp.Status)
	}
}

func TestGetStreamingInfo_CacheMiss_StorageError_ReturnsNotUploaded(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)

	store := &mockStorage{
		listFunc: func(ctx context.Context, bucket, prefix string) ([]storage.ObjectInfo, error) {
			return nil, errors.New("storage unavailable")
		},
	}

	server := NewStreamingServer(store, svc, client, config.MinIOConfig{
		EncodedBucket: "streamvault-encoded",
	})

	resp, err := server.GetStreamingInfo(context.Background(), &streamingv1.GetStreamingInfoRequest{
		ContentId: "movie-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != commonv1.VideoStatus_VIDEO_STATUS_NOT_UPLOADED {
		t.Errorf("expected NOT_UPLOADED on storage error, got %v", resp.Status)
	}
}

// --- GetProgress tests ---

func TestGetProgress_EmptyUserId_ReturnsInvalidArgument(t *testing.T) {
	server, _, _ := setupServer(t)

	_, err := server.GetProgress(context.Background(), &streamingv1.GetProgressRequest{
		UserId:    "",
		ContentId: "movie-1",
	})

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestGetProgress_EmptyContentId_ReturnsInvalidArgument(t *testing.T) {
	server, _, _ := setupServer(t)

	_, err := server.GetProgress(context.Background(), &streamingv1.GetProgressRequest{
		UserId:    "user-1",
		ContentId: "",
	})

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestGetProgress_NotFound_ReturnsNotFound(t *testing.T) {
	server, _, _ := setupServer(t)

	_, err := server.GetProgress(context.Background(), &streamingv1.GetProgressRequest{
		UserId:    "user-1",
		ContentId: "nonexistent",
	})

	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestGetProgress_Found_ReturnsCorrectData(t *testing.T) {
	server, client, _ := setupServer(t)
	ctx := context.Background()

	// Save progress via Redis directly
	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)
	_ = svc.SaveProgress(ctx, "user-1", "movie-1", 60, 120)

	// Re-create server with the same progress service
	server.progressSvc = svc

	resp, err := server.GetProgress(ctx, &streamingv1.GetProgressRequest{
		UserId:    "user-1",
		ContentId: "movie-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserId != "user-1" {
		t.Errorf("expected user-1, got %s", resp.UserId)
	}
	if resp.ContentId != "movie-1" {
		t.Errorf("expected movie-1, got %s", resp.ContentId)
	}
	if resp.PositionSeconds != 60 {
		t.Errorf("expected position 60, got %d", resp.PositionSeconds)
	}
	if resp.DurationSeconds != 120 {
		t.Errorf("expected duration 120, got %d", resp.DurationSeconds)
	}
	if resp.Percentage != 50.0 {
		t.Errorf("expected percentage 50.0, got %f", resp.Percentage)
	}
}

// --- GetContinueWatching tests ---

func TestGetContinueWatching_EmptyUserId_ReturnsInvalidArgument(t *testing.T) {
	server, _, _ := setupServer(t)

	_, err := server.GetContinueWatching(context.Background(), &streamingv1.GetContinueWatchingRequest{
		UserId: "",
	})

	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestGetContinueWatching_DefaultLimit(t *testing.T) {
	server, client, _ := setupServer(t)
	ctx := context.Background()

	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)
	server.progressSvc = svc

	_ = svc.SaveProgress(ctx, "user-1", "movie-1", 30, 120)

	resp, err := server.GetContinueWatching(ctx, &streamingv1.GetContinueWatchingRequest{
		UserId: "user-1",
		Limit:  0, // should default to 20
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestGetContinueWatching_CustomLimit(t *testing.T) {
	server, client, _ := setupServer(t)
	ctx := context.Background()

	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)
	server.progressSvc = svc

	_ = svc.SaveProgress(ctx, "user-1", "movie-1", 30, 120)
	_ = svc.SaveProgress(ctx, "user-1", "movie-2", 60, 120)
	_ = svc.SaveProgress(ctx, "user-1", "movie-3", 10, 120)

	resp, err := server.GetContinueWatching(ctx, &streamingv1.GetContinueWatchingRequest{
		UserId: "user-1",
		Limit:  2,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items with limit=2, got %d", len(resp.Items))
	}
}

func TestGetContinueWatching_ReturnsCorrectFields(t *testing.T) {
	server, client, _ := setupServer(t)
	ctx := context.Background()

	repo := progress.NewRedisRepository(client, 90*24*time.Hour)
	svc := progress.NewService(repo, nil)
	server.progressSvc = svc

	_ = svc.SaveProgress(ctx, "user-1", "movie-1", 60, 120)

	resp, err := server.GetContinueWatching(ctx, &streamingv1.GetContinueWatchingRequest{
		UserId: "user-1",
		Limit:  10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}

	item := resp.Items[0]
	if item.ContentId != "movie-1" {
		t.Errorf("expected movie-1, got %s", item.ContentId)
	}
	if item.PositionSeconds != 60 {
		t.Errorf("expected position 60, got %d", item.PositionSeconds)
	}
	if item.DurationSeconds != 120 {
		t.Errorf("expected duration 120, got %d", item.DurationSeconds)
	}
	if item.Percentage != 50.0 {
		t.Errorf("expected percentage 50.0, got %f", item.Percentage)
	}
	if item.UpdatedAt == nil {
		t.Error("expected UpdatedAt to be set")
	}
}

// --- qualityToMinTier tests ---

func TestQualityToMinTier(t *testing.T) {
	tests := []struct {
		quality string
		want    commonv1.SubscriptionTier
	}{
		{"360p", commonv1.SubscriptionTier_SUBSCRIPTION_TIER_BASIC},
		{"720p", commonv1.SubscriptionTier_SUBSCRIPTION_TIER_BASIC},
		{"1080p", commonv1.SubscriptionTier_SUBSCRIPTION_TIER_STANDARD},
		{"4k", commonv1.SubscriptionTier_SUBSCRIPTION_TIER_PREMIUM},
		{"unknown", commonv1.SubscriptionTier_SUBSCRIPTION_TIER_BASIC},
	}

	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			got := qualityToMinTier(tt.quality)
			if got != tt.want {
				t.Errorf("qualityToMinTier(%q) = %v, want %v", tt.quality, got, tt.want)
			}
		})
	}
}
