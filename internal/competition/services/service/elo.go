package service

import (
	"math"
)

const (
	// KFactor is the maximum adjustment per game
	KFactor = 32
)

// CalculateELO calculates the new ratings for two bots based on the outcome of a match.
// scoreA: 1 if Bot A wins, 0.5 for draw, 0 if Bot B wins.
func CalculateELO(ratingA, ratingB float64, scoreA float64) (newA, newB float64) {
	// Expected scores
	expectedA := 1 / (1 + math.Pow(10, (ratingB-ratingA)/400))
	expectedB := 1 / (1 + math.Pow(10, (ratingA-ratingB)/400))

	// New ratings
	newA = ratingA + KFactor*(scoreA-expectedA)
	newB = ratingB + KFactor*((1-scoreA)-expectedB)

	return newA, newB
}

// CalculateMultiELO calculates ratings for multiple bots (e.g., Battle Royale)
// This is a simplified version where we treat it as series of 1v1s against average rank
func CalculateMultiELO(ratings []float64, winnerIndex int) []float64 {
	newRatings := make([]float64, len(ratings))
	copy(newRatings, ratings)

	for i := range ratings {
		if i == winnerIndex {
			continue
		}
		// Winner vs each loser
		newWinner, newLoser := CalculateELO(newRatings[winnerIndex], newRatings[i], 1.0)
		newRatings[winnerIndex] = newWinner
		newRatings[i] = newLoser
	}

	return newRatings
}
