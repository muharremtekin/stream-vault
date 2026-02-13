package progress

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupRepo(t *testing.T) (*miniredis.Miniredis, Repository) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := NewRedisRepository(client, 90*24*time.Hour)
	return mr, repo
}

func TestRepository_SaveAndGet(t *testing.T) {
	_, repo := setupRepo(t)
	ctx := context.Background()

	p := WatchProgress{
		UserID:          "user-1",
		ContentID:       "movie-1",
		PositionSeconds: 300,
		DurationSeconds: 7200,
		Percentage:      4.16,
		UpdatedAt:       time.Now().Unix(),
	}

	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := repo.Get(ctx, "user-1", "movie-1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil, expected progress")
	}
	if got.PositionSeconds != 300 {
		t.Errorf("PositionSeconds = %d, want 300", got.PositionSeconds)
	}
	if got.DurationSeconds != 7200 {
		t.Errorf("DurationSeconds = %d, want 7200", got.DurationSeconds)
	}
}

func TestRepository_GetNonExistent(t *testing.T) {
	_, repo := setupRepo(t)
	ctx := context.Background()

	got, err := repo.Get(ctx, "user-1", "nonexistent")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != nil {
		t.Error("Get() should return nil for non-existent progress")
	}
}

func TestRepository_ContinueWatching_TTLSet(t *testing.T) {
	mr, repo := setupRepo(t)
	ctx := context.Background()

	p := WatchProgress{
		UserID:          "user-1",
		ContentID:       "movie-1",
		PositionSeconds: 100,
		DurationSeconds: 7200,
		Percentage:      1.38,
		UpdatedAt:       time.Now().Unix(),
	}

	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	cwKey := "continue-watching:user-1"
	ttl := mr.TTL(cwKey)
	if ttl <= 0 {
		t.Errorf("continue-watching key should have TTL, got %v", ttl)
	}
	// TTL should be approximately 90 days (allow 1 minute tolerance)
	expected := 90 * 24 * time.Hour
	if ttl < expected-time.Minute || ttl > expected+time.Minute {
		t.Errorf("TTL = %v, want approximately %v", ttl, expected)
	}
}

func TestRepository_ContinueWatching_CappedAt100(t *testing.T) {
	mr, repo := setupRepo(t)
	ctx := context.Background()

	// Save 110 entries
	for i := 0; i < 110; i++ {
		p := WatchProgress{
			UserID:          "user-1",
			ContentID:       fmt.Sprintf("movie-%d", i),
			PositionSeconds: 100,
			DurationSeconds: 7200,
			Percentage:      1.38,
			UpdatedAt:       int64(1000000 + i), // increasing timestamps
		}
		if err := repo.Save(ctx, p); err != nil {
			t.Fatalf("Save() error for movie-%d: %v", i, err)
		}
	}

	// Verify sorted set has exactly 100 members
	cwKey := "continue-watching:user-1"
	members, err := mr.ZMembers(cwKey)
	if err != nil {
		t.Fatalf("ZMembers() error: %v", err)
	}
	if len(members) != 100 {
		t.Errorf("continue-watching should have 100 members, got %d", len(members))
	}

	// Verify oldest entries were removed (movie-0 through movie-9)
	memberSet := make(map[string]bool)
	for _, m := range members {
		memberSet[m] = true
	}
	for i := 0; i < 10; i++ {
		contentID := fmt.Sprintf("movie-%d", i)
		if memberSet[contentID] {
			t.Errorf("old entry %s should have been removed", contentID)
		}
	}

	// Verify newest entries are still present
	for i := 10; i < 110; i++ {
		contentID := fmt.Sprintf("movie-%d", i)
		if !memberSet[contentID] {
			t.Errorf("recent entry %s should be present", contentID)
		}
	}
}

func TestRepository_GetContinueWatching_OrderedByMostRecent(t *testing.T) {
	_, repo := setupRepo(t)
	ctx := context.Background()

	entries := []WatchProgress{
		{UserID: "user-1", ContentID: "movie-old", PositionSeconds: 50, DurationSeconds: 7200, Percentage: 0.69, UpdatedAt: 1000},
		{UserID: "user-1", ContentID: "movie-mid", PositionSeconds: 100, DurationSeconds: 7200, Percentage: 1.38, UpdatedAt: 2000},
		{UserID: "user-1", ContentID: "movie-new", PositionSeconds: 200, DurationSeconds: 7200, Percentage: 2.77, UpdatedAt: 3000},
	}

	for _, p := range entries {
		if err := repo.Save(ctx, p); err != nil {
			t.Fatalf("Save() error: %v", err)
		}
	}

	items, err := repo.GetContinueWatching(ctx, "user-1", 10)
	if err != nil {
		t.Fatalf("GetContinueWatching() error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	if items[0].ContentID != "movie-new" {
		t.Errorf("first item should be movie-new (most recent), got %s", items[0].ContentID)
	}
	if items[1].ContentID != "movie-mid" {
		t.Errorf("second item should be movie-mid, got %s", items[1].ContentID)
	}
	if items[2].ContentID != "movie-old" {
		t.Errorf("third item should be movie-old, got %s", items[2].ContentID)
	}
}

func TestRepository_GetContinueWatching_DefaultLimit(t *testing.T) {
	_, repo := setupRepo(t)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		p := WatchProgress{
			UserID:          "user-1",
			ContentID:       fmt.Sprintf("movie-%d", i),
			PositionSeconds: 100,
			DurationSeconds: 7200,
			Percentage:      1.38,
			UpdatedAt:       int64(1000 + i),
		}
		if err := repo.Save(ctx, p); err != nil {
			t.Fatalf("Save() error: %v", err)
		}
	}

	// limit=0 should default to 20
	items, err := repo.GetContinueWatching(ctx, "user-1", 0)
	if err != nil {
		t.Fatalf("GetContinueWatching() error: %v", err)
	}
	if len(items) != 20 {
		t.Errorf("expected 20 items with default limit, got %d", len(items))
	}
}

func TestRepository_Remove(t *testing.T) {
	mr, repo := setupRepo(t)
	ctx := context.Background()

	p := WatchProgress{
		UserID:          "user-1",
		ContentID:       "movie-1",
		PositionSeconds: 100,
		DurationSeconds: 7200,
		Percentage:      1.38,
		UpdatedAt:       time.Now().Unix(),
	}

	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := repo.Remove(ctx, "user-1", "movie-1"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	got, err := repo.Get(ctx, "user-1", "movie-1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != nil {
		t.Error("progress should be nil after Remove()")
	}

	// Key may not exist if ZRem removed the only member (miniredis deletes empty sets)
	if mr.Exists("continue-watching:user-1") {
		members, _ := mr.ZMembers("continue-watching:user-1")
		for _, m := range members {
			if m == "movie-1" {
				t.Error("movie-1 should be removed from continue-watching set")
			}
		}
	}
}

func TestRepository_ProgressHashTTL(t *testing.T) {
	mr, repo := setupRepo(t)
	ctx := context.Background()

	p := WatchProgress{
		UserID:          "user-1",
		ContentID:       "movie-1",
		PositionSeconds: 100,
		DurationSeconds: 7200,
		Percentage:      1.38,
		UpdatedAt:       time.Now().Unix(),
	}

	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	key := "progress:user-1:movie-1"
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Errorf("progress key should have TTL, got %v", ttl)
	}
}
