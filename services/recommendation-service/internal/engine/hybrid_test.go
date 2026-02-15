package engine

import "testing"

func TestDynamicWeights_ColdStart(t *testing.T) {
	alpha, beta, gamma := dynamicWeights(3)
	if alpha != 0.0 || beta != 0.1 || gamma != 0.9 {
		t.Errorf("cold start (<5): expected (0, 0.1, 0.9), got (%f, %f, %f)", alpha, beta, gamma)
	}
}

func TestDynamicWeights_Medium(t *testing.T) {
	alpha, beta, gamma := dynamicWeights(10)
	if alpha != 0.2 || beta != 0.5 || gamma != 0.3 {
		t.Errorf("medium (10): expected (0.2, 0.5, 0.3), got (%f, %f, %f)", alpha, beta, gamma)
	}
}

func TestDynamicWeights_High(t *testing.T) {
	alpha, beta, gamma := dynamicWeights(30)
	if alpha != 0.4 || beta != 0.4 || gamma != 0.2 {
		t.Errorf("high (30): expected (0.4, 0.4, 0.2), got (%f, %f, %f)", alpha, beta, gamma)
	}
}

func TestDynamicWeights_VeryHigh(t *testing.T) {
	alpha, beta, gamma := dynamicWeights(60)
	if alpha != 0.5 || beta != 0.35 || gamma != 0.15 {
		t.Errorf("very high (60): expected (0.5, 0.35, 0.15), got (%f, %f, %f)", alpha, beta, gamma)
	}
}

func TestDominantAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		collab   float64
		content  float64
		pop      float64
		expected string
	}{
		{"clear collaborative winner", 5.0, 1.0, 1.0, "collaborative"},
		{"clear content_based winner", 1.0, 5.0, 1.0, "content_based"},
		{"clear popularity winner", 1.0, 1.0, 5.0, "popularity"},
		{"close scores → hybrid", 4.0, 3.5, 1.0, "hybrid"},
		{"all equal → hybrid", 3.0, 3.0, 3.0, "hybrid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dominantAlgorithm(tt.collab, tt.content, tt.pop)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
