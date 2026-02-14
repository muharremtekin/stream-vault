package progress

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupServiceTest(t *testing.T) (*Service, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := NewRedisRepository(client, 90*24*time.Hour)
	svc := NewService(repo, nil)
	return svc, mr
}

func TestService_SaveProgress_NormalProgress(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	err := svc.SaveProgress(ctx, "user-1", "movie-1", 60, 120)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := svc.GetProgress(ctx, "user-1", "movie-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected progress to be saved")
	}
	if p.PositionSeconds != 60 {
		t.Errorf("expected position 60, got %d", p.PositionSeconds)
	}
	if p.DurationSeconds != 120 {
		t.Errorf("expected duration 120, got %d", p.DurationSeconds)
	}
	if p.Percentage != 50.0 {
		t.Errorf("expected percentage 50.0, got %f", p.Percentage)
	}

	// Should also appear in continue-watching
	items, err := svc.GetContinueWatching(ctx, "user-1", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 continue-watching item, got %d", len(items))
	}
}

func TestService_SaveProgress_AtCompletionThreshold_RemovesFromList(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	// First save some progress
	err := svc.SaveProgress(ctx, "user-1", "movie-1", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it exists
	items, _ := svc.GetContinueWatching(ctx, "user-1", 20)
	if len(items) != 1 {
		t.Fatalf("expected 1 item before completion, got %d", len(items))
	}

	// Now save at exactly 95% (completion threshold)
	err = svc.SaveProgress(ctx, "user-1", "movie-1", 95, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be removed from continue-watching
	items, _ = svc.GetContinueWatching(ctx, "user-1", 20)
	if len(items) != 0 {
		t.Errorf("expected 0 items after completion, got %d", len(items))
	}
}

func TestService_SaveProgress_AboveThreshold_RemovesFromList(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	// Save initial progress
	_ = svc.SaveProgress(ctx, "user-1", "movie-1", 50, 100)

	// Save at 96%
	err := svc.SaveProgress(ctx, "user-1", "movie-1", 96, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, _ := svc.GetContinueWatching(ctx, "user-1", 20)
	if len(items) != 0 {
		t.Errorf("expected 0 items after 96%% completion, got %d", len(items))
	}
}

func TestService_SaveProgress_BelowThreshold_KeepsInList(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	// Save at 94% (just below 95% threshold)
	err := svc.SaveProgress(ctx, "user-1", "movie-1", 94, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, _ := svc.GetContinueWatching(ctx, "user-1", 20)
	if len(items) != 1 {
		t.Errorf("expected 1 item at 94%%, got %d", len(items))
	}
}

func TestService_SaveProgress_ZeroDuration_NoPanic(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	// Duration <= 0 is treated as 1 → percentage = position/1 * 100
	// With position=50, that's 5000% which exceeds 95% threshold → triggers Remove
	err := svc.SaveProgress(ctx, "user-1", "movie-1", 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Progress should NOT exist because 5000% >= 95% triggers removal
	p, _ := svc.GetProgress(ctx, "user-1", "movie-1")
	if p != nil {
		t.Error("expected nil progress since zero duration causes completion threshold to trigger")
	}
}

func TestService_SaveProgress_NegativeDuration_NoPanic(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	// Negative duration is also treated as 1
	err := svc.SaveProgress(ctx, "user-1", "movie-1", 0, -5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Position=0, duration=1 → 0% → saved normally
	p, _ := svc.GetProgress(ctx, "user-1", "movie-1")
	if p == nil {
		t.Fatal("expected progress to be saved with position=0")
	}
	if p.Percentage != 0.0 {
		t.Errorf("expected percentage 0.0, got %f", p.Percentage)
	}
}

func TestService_SaveProgress_ZeroPosition(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	err := svc.SaveProgress(ctx, "user-1", "movie-1", 0, 120)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, _ := svc.GetProgress(ctx, "user-1", "movie-1")
	if p == nil {
		t.Fatal("expected progress to be saved")
	}
	if p.Percentage != 0.0 {
		t.Errorf("expected percentage 0.0, got %f", p.Percentage)
	}
}

func TestService_GetProgress_NonExistent_ReturnsNil(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	p, err := svc.GetProgress(ctx, "user-1", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != nil {
		t.Error("expected nil for non-existent progress")
	}
}

func TestService_GetContinueWatching_Empty(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	items, err := svc.GetContinueWatching(ctx, "user-1", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestService_GetContinueWatching_MultipleItems(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	_ = svc.SaveProgress(ctx, "user-1", "movie-1", 30, 120)
	_ = svc.SaveProgress(ctx, "user-1", "movie-2", 60, 120)
	_ = svc.SaveProgress(ctx, "user-1", "movie-3", 10, 120)

	items, err := svc.GetContinueWatching(ctx, "user-1", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestService_GetContinueWatching_RespectsLimit(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	_ = svc.SaveProgress(ctx, "user-1", "movie-1", 30, 120)
	_ = svc.SaveProgress(ctx, "user-1", "movie-2", 60, 120)
	_ = svc.SaveProgress(ctx, "user-1", "movie-3", 10, 120)

	items, err := svc.GetContinueWatching(ctx, "user-1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items with limit=2, got %d", len(items))
	}
}
