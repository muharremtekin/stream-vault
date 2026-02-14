package engine

import "math"

// CosineSimilarity computes cosine similarity between two sparse vectors
// represented as map[string]float64. Returns 0.0 if either vector is zero.
func CosineSimilarity(a, b map[string]float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	dot := dotProduct(a, b)
	magA := magnitude(a)
	magB := magnitude(b)

	if magA == 0 || magB == 0 {
		return 0.0
	}
	return dot / (magA * magB)
}

// dotProduct computes the dot product of two sparse vectors.
// Iterates over the smaller map for efficiency.
func dotProduct(a, b map[string]float64) float64 {
	// Iterate over the smaller map
	if len(a) > len(b) {
		a, b = b, a
	}

	var sum float64
	for key, valA := range a {
		if valB, ok := b[key]; ok {
			sum += valA * valB
		}
	}
	return sum
}

// magnitude computes the L2 norm (magnitude) of a sparse vector.
func magnitude(v map[string]float64) float64 {
	var sum float64
	for _, val := range v {
		sum += val * val
	}
	return math.Sqrt(sum)
}
