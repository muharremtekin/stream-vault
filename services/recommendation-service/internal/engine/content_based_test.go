package engine

import (
	"math"
	"testing"
	"time"

	"github.com/streamvault/recommendation-service/internal/model"
)

func TestBuildFeatureVector_AllFields(t *testing.T) {
	cf := &model.ContentFeatures{
		ContentID:   "c1",
		Title:       "Test Movie",
		Genres:      []string{"Action", "Drama"},
		Tags:        []string{"epic", "war"},
		Director:    "Christopher Nolan",
		ReleaseYear: 2014,
		AvgRating:   8.5,
	}

	vec := BuildFeatureVector(cf)

	// Check genre keys
	if vec["genre:action"] != 1.0 {
		t.Errorf("expected genre:action = 1.0, got %f", vec["genre:action"])
	}
	if vec["genre:drama"] != 1.0 {
		t.Errorf("expected genre:drama = 1.0, got %f", vec["genre:drama"])
	}

	// Check tag keys
	if vec["tag:epic"] != 1.0 {
		t.Errorf("expected tag:epic = 1.0, got %f", vec["tag:epic"])
	}
	if vec["tag:war"] != 1.0 {
		t.Errorf("expected tag:war = 1.0, got %f", vec["tag:war"])
	}

	// Check director key
	if vec["director:christopher-nolan"] != 1.0 {
		t.Errorf("expected director:christopher-nolan = 1.0, got %f", vec["director:christopher-nolan"])
	}

	// Check rating normalized
	if !almostEqual(vec["rating"], 0.85, 1e-9) {
		t.Errorf("expected rating = 0.85, got %f", vec["rating"])
	}

	// Check year is present and normalized
	if _, ok := vec["year"]; !ok {
		t.Error("expected year key in feature vector")
	}
}

func TestBuildFeatureVector_EmptyContent(t *testing.T) {
	cf := &model.ContentFeatures{
		ContentID:   "c2",
		ReleaseYear: 2000,
	}

	vec := BuildFeatureVector(cf)

	// Should have year but not genres, tags, director, or rating
	if _, ok := vec["year"]; !ok {
		t.Error("expected year key even for empty content")
	}
	if _, ok := vec["rating"]; ok {
		t.Error("did not expect rating key for zero rating")
	}

	// No genre, tag, director keys
	for k := range vec {
		if k != "year" {
			t.Errorf("unexpected key %q in empty content vector", k)
		}
	}
}

func TestBuildFeatureVector_YearNormalization(t *testing.T) {
	cf := &model.ContentFeatures{
		ContentID:   "c3",
		ReleaseYear: 2000,
	}

	vec := BuildFeatureVector(cf)

	currentYear := time.Now().Year()
	expected := float64(2000-minYear) / float64(currentYear-minYear)
	if !almostEqual(vec["year"], expected, 1e-6) {
		t.Errorf("expected year = %f, got %f", expected, vec["year"])
	}
}

func TestBuildFeatureVector_RatingNormalization(t *testing.T) {
	cf := &model.ContentFeatures{
		ContentID:   "c4",
		ReleaseYear: 2020,
		AvgRating:   7.0,
	}

	vec := BuildFeatureVector(cf)

	if !almostEqual(vec["rating"], 0.7, 1e-9) {
		t.Errorf("expected rating = 0.7, got %f", vec["rating"])
	}
}

func TestBuildUserPreferenceVector_SingleInteraction(t *testing.T) {
	cf := &model.ContentFeatures{
		ContentID:   "c1",
		Genres:      []string{"Action"},
		ReleaseYear: 2020,
		AvgRating:   8.0,
		FeatureVector: map[string]float64{
			"genre:action": 1.0,
			"rating":       0.8,
		},
	}

	interactions := []model.Interaction{
		{
			UserID:          "u1",
			ContentID:       "c1",
			InteractionType: model.InteractionComplete,
		},
	}
	contentMap := map[string]*model.ContentFeatures{"c1": cf}

	pref := BuildUserPreferenceVector(interactions, contentMap)
	if pref == nil {
		t.Fatal("expected non-nil preference vector")
	}

	// Complete interaction has weight 1.0, single interaction means pref = feature vector
	if !almostEqual(pref["genre:action"], 1.0, 1e-9) {
		t.Errorf("expected genre:action = 1.0, got %f", pref["genre:action"])
	}
	if !almostEqual(pref["rating"], 0.8, 1e-9) {
		t.Errorf("expected rating = 0.8, got %f", pref["rating"])
	}
}

func TestBuildUserPreferenceVector_MultipleInteractions(t *testing.T) {
	cf1 := &model.ContentFeatures{
		ContentID: "c1",
		FeatureVector: map[string]float64{
			"genre:action": 1.0,
			"rating":       0.8,
		},
	}
	cf2 := &model.ContentFeatures{
		ContentID: "c2",
		FeatureVector: map[string]float64{
			"genre:comedy": 1.0,
			"rating":       0.6,
		},
	}

	interactions := []model.Interaction{
		{UserID: "u1", ContentID: "c1", InteractionType: model.InteractionComplete}, // weight 1.0
		{UserID: "u1", ContentID: "c2", InteractionType: model.InteractionBookmark}, // weight 0.6
	}
	contentMap := map[string]*model.ContentFeatures{"c1": cf1, "c2": cf2}

	pref := BuildUserPreferenceVector(interactions, contentMap)
	if pref == nil {
		t.Fatal("expected non-nil preference vector")
	}

	// Total weight = 1.0 + 0.6 = 1.6
	// genre:action = 1.0 * 1.0 / 1.6 = 0.625
	// genre:comedy = 1.0 * 0.6 / 1.6 = 0.375
	// rating = (0.8*1.0 + 0.6*0.6) / 1.6 = (0.8 + 0.36) / 1.6 = 0.725
	if !almostEqual(pref["genre:action"], 0.625, 1e-3) {
		t.Errorf("expected genre:action ≈ 0.625, got %f", pref["genre:action"])
	}
	if !almostEqual(pref["genre:comedy"], 0.375, 1e-3) {
		t.Errorf("expected genre:comedy ≈ 0.375, got %f", pref["genre:comedy"])
	}
	if !almostEqual(pref["rating"], 0.725, 1e-3) {
		t.Errorf("expected rating ≈ 0.725, got %f", pref["rating"])
	}
}

func TestBuildUserPreferenceVector_EmptyInteractions(t *testing.T) {
	pref := BuildUserPreferenceVector(nil, nil)
	if pref != nil {
		t.Error("expected nil for empty interactions")
	}

	pref = BuildUserPreferenceVector([]model.Interaction{}, map[string]*model.ContentFeatures{})
	if pref != nil {
		t.Error("expected nil for empty interactions slice")
	}
}

func TestScoreContent_ReturnsCosineSimilarity(t *testing.T) {
	userPref := map[string]float64{
		"genre:action": 1.0,
		"rating":       0.8,
	}
	cf := &model.ContentFeatures{
		ContentID: "c1",
		FeatureVector: map[string]float64{
			"genre:action": 1.0,
			"genre:drama":  1.0,
			"rating":       0.9,
		},
	}

	score := ScoreContent(userPref, cf)
	expected := CosineSimilarity(userPref, cf.FeatureVector)

	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("ScoreContent (%f) does not match CosineSimilarity (%f)", score, expected)
	}
}
