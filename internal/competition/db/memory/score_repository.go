package memory

import (
	"sync"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
)

type InMemoryScoreRepository struct {
	mu     sync.RWMutex
	scores map[string]int // participantID -> total score
}

func NewInMemoryScoreRepository() *InMemoryScoreRepository {
	return &InMemoryScoreRepository{
		scores: make(map[string]int),
	}
}

func (r *InMemoryScoreRepository) UpdateScore(matchID string, participantID string, points int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scores[participantID] += points
	return nil
}

func (r *InMemoryScoreRepository) GetLeaderboard(challengeID string) ([]*entity.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Mock implementation: convert map to list
	var leaderboard []*entity.Participant
	for pid, score := range r.scores {
		leaderboard = append(leaderboard, &entity.Participant{
			ID:    pid,
			Score: score,
		})
	}
	return leaderboard, nil
}
