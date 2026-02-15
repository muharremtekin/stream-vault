package engine

import (
	"testing"

	"github.com/streamvault/recommendation-service/internal/model"
)

func TestImplicitRating_Rate(t *testing.T) {
	rating := 8.0
	i := &model.Interaction{InteractionType: model.InteractionRate, Rating: &rating}
	got := implicitRating(i)
	if got != 8.0 {
		t.Errorf("expected 8.0, got %f", got)
	}
}

func TestImplicitRating_Complete(t *testing.T) {
	i := &model.Interaction{InteractionType: model.InteractionComplete}
	got := implicitRating(i)
	if got != 7.0 {
		t.Errorf("expected 7.0, got %f", got)
	}
}

func TestImplicitRating_Bookmark(t *testing.T) {
	i := &model.Interaction{InteractionType: model.InteractionBookmark}
	got := implicitRating(i)
	if got != 6.0 {
		t.Errorf("expected 6.0, got %f", got)
	}
}

func TestImplicitRating_ViewCompletionLevels(t *testing.T) {
	tests := []struct {
		name     string
		pct      float64
		expected float64
	}{
		{"95% completion", 95, 7.0},
		{"60% completion", 60, 5.0},
		{"30% completion", 30, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &model.Interaction{
				InteractionType: model.InteractionView,
				CompletionPct:   tt.pct,
			}
			got := implicitRating(i)
			if got != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, got)
			}
		})
	}
}

func TestFindSimilarUsers_ReturnsTopK(t *testing.T) {
	matrix := map[string]map[string]float64{
		"target": {"c1": 5.0, "c2": 3.0},
		"u1":     {"c1": 5.0, "c2": 3.0},                  // identical to target → sim ≈ 1.0
		"u2":     {"c1": 1.0, "c2": 1.0},                  // different direction
		"u3":     {"c1": 4.0, "c2": 2.5},                  // similar to target
		"u4":     {"c3": 5.0, "c4": 3.0},                  // no overlap → sim = 0
	}

	result := findSimilarUsers("target", matrix, 2)

	if len(result) != 2 {
		t.Fatalf("expected 2 similar users, got %d", len(result))
	}

	// Results should be sorted by similarity descending
	if result[0].Similarity < result[1].Similarity {
		t.Error("expected results sorted by similarity descending")
	}

	// u1 should be the most similar (identical direction)
	if result[0].UserID != "u1" {
		t.Errorf("expected u1 as most similar, got %s", result[0].UserID)
	}
}

func TestPredictRating_WeightedAverage(t *testing.T) {
	similarUsers := []userSimilarity{
		{UserID: "u1", Similarity: 0.9},
		{UserID: "u2", Similarity: 0.6},
	}
	matrix := map[string]map[string]float64{
		"u1": {"c1": 8.0},
		"u2": {"c1": 6.0},
	}

	predicted := predictRating("c1", similarUsers, matrix)
	// weighted = 0.9*8.0 + 0.6*6.0 = 7.2 + 3.6 = 10.8
	// simSum = 0.9 + 0.6 = 1.5
	// predicted = 10.8 / 1.5 = 7.2
	expected := 7.2
	if !almostEqual(predicted, expected, 1e-9) {
		t.Errorf("expected %f, got %f", expected, predicted)
	}
}
