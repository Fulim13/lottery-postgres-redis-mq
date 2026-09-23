package util

import "math/rand/v2"

// Binary search over acc, the ascending array of cumulative probabilities: return the index
// of the first entry where acc[i] >= target, i.e. which segment target falls into.
// For a non-empty acc the result is always within [0, len(acc)-1].
func BinarySearch4Section(acc []float64, target float64) int {
	low, high := 0, len(acc)-1
	for low < high {
		mid := low + (high-low)/2
		if acc[mid] < target {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

// Draw a prize. Takes the odds of each prize (they need not be normalized, but every one of
// them must be greater than zero) and returns the index of the prize that was drawn.
func Lottery(probs []float64) int {
	if len(probs) == 0 {
		return -1
	}
	sum := 0.0
	acc := make([]float64, 0, len(probs)) // cumulative probabilities
	for _, prob := range probs {
		sum += prob
		acc = append(acc, sum)
	}

	// a random number in (0, sum]
	r := rand.Float64() * sum
	index := BinarySearch4Section(acc, r)
	return index
}
