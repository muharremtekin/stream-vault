package engine

import (
	"math"
	"testing"
)

func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

func TestCosineSimilarity_IdenticalVectors(t *testing.T) {
	v := map[string]float64{"a": 1.0, "b": 2.0, "c": 3.0}
	sim := CosineSimilarity(v, v)
	if !almostEqual(sim, 1.0, 1e-9) {
		t.Errorf("expected 1.0 for identical vectors, got %f", sim)
	}
}

func TestCosineSimilarity_OrthogonalVectors(t *testing.T) {
	a := map[string]float64{"x": 1.0, "y": 0.0}
	b := map[string]float64{"z": 1.0, "w": 2.0}
	sim := CosineSimilarity(a, b)
	if sim != 0.0 {
		t.Errorf("expected 0.0 for orthogonal vectors, got %f", sim)
	}
}

func TestCosineSimilarity_KnownValues(t *testing.T) {
	a := map[string]float64{"a": 1.0, "b": 2.0}
	b := map[string]float64{"a": 2.0, "b": 1.0}
	// dot = 1*2 + 2*1 = 4
	// magA = sqrt(1+4) = sqrt(5)
	// magB = sqrt(4+1) = sqrt(5)
	// cos = 4/5 = 0.8
	expected := 0.8
	sim := CosineSimilarity(a, b)
	if !almostEqual(sim, expected, 1e-9) {
		t.Errorf("expected %f, got %f", expected, sim)
	}
}

func TestCosineSimilarity_EmptyVectors(t *testing.T) {
	a := map[string]float64{"a": 1.0}
	empty := map[string]float64{}

	if sim := CosineSimilarity(a, empty); sim != 0.0 {
		t.Errorf("expected 0.0 for empty second vector, got %f", sim)
	}
	if sim := CosineSimilarity(empty, a); sim != 0.0 {
		t.Errorf("expected 0.0 for empty first vector, got %f", sim)
	}
	if sim := CosineSimilarity(nil, nil); sim != 0.0 {
		t.Errorf("expected 0.0 for nil vectors, got %f", sim)
	}
}

func TestCosineSimilarity_SingleDimension(t *testing.T) {
	a := map[string]float64{"x": 3.0}
	b := map[string]float64{"x": 7.0}
	// Same direction → cos = 1.0
	sim := CosineSimilarity(a, b)
	if !almostEqual(sim, 1.0, 1e-9) {
		t.Errorf("expected 1.0 for parallel vectors, got %f", sim)
	}
}

func TestCosineSimilarity_PartialOverlap(t *testing.T) {
	a := map[string]float64{"x": 1.0, "y": 1.0, "z": 0.0}
	b := map[string]float64{"x": 1.0, "w": 1.0}
	// dot = 1*1 = 1
	// magA = sqrt(1+1+0) = sqrt(2)
	// magB = sqrt(1+1) = sqrt(2)
	// cos = 1/2 = 0.5
	sim := CosineSimilarity(a, b)
	if !almostEqual(sim, 0.5, 1e-9) {
		t.Errorf("expected 0.5, got %f", sim)
	}
}
