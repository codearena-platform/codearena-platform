package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateMultiELO(t *testing.T) {
	t.Run("winner gains and loser loses", func(t *testing.T) {
		ratings := []float64{1200, 1200}
		updated := CalculateMultiELO(ratings, 0)

		assert.Greater(t, updated[0], ratings[0])
		assert.Less(t, updated[1], ratings[1])
		assert.InDelta(t, 2400, updated[0]+updated[1], 0.001)
		assert.Equal(t, []float64{1200, 1200}, ratings, "input must not be mutated")
	})

	t.Run("invalid winner index leaves ratings unchanged", func(t *testing.T) {
		for _, winner := range []int{-1, 2, 99} {
			ratings := []float64{1100, 1300}
			assert.Equal(t, ratings, CalculateMultiELO(ratings, winner), "winner %d", winner)
		}
	})

	t.Run("empty ratings are safe", func(t *testing.T) {
		assert.Empty(t, CalculateMultiELO(nil, 0))
	})
}

func TestCalculateMultiELODraw(t *testing.T) {
	t.Run("equal ratings stay level", func(t *testing.T) {
		assert.Equal(t, []float64{1200, 1200, 1200}, CalculateMultiELODraw([]float64{1200, 1200, 1200}))
	})

	t.Run("unequal ratings converge without changing total", func(t *testing.T) {
		ratings := []float64{1000, 1400}
		updated := CalculateMultiELODraw(ratings)

		assert.Greater(t, updated[0], ratings[0])
		assert.Less(t, updated[1], ratings[1])
		assert.InDelta(t, 2400, updated[0]+updated[1], 0.001)
	})

	t.Run("empty ratings are safe", func(t *testing.T) {
		assert.Empty(t, CalculateMultiELODraw(nil))
	})
}
