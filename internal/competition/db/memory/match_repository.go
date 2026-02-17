package memory

import (
	"errors"
	"sync"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
)

type InMemoryMatchRepository struct {
	mu      sync.RWMutex
	matches map[string]*entity.Match
}

func NewInMemoryMatchRepository() *InMemoryMatchRepository {
	return &InMemoryMatchRepository{
		matches: make(map[string]*entity.Match),
	}
}

func (r *InMemoryMatchRepository) Save(match *entity.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[match.ID] = match
	return nil
}

func (r *InMemoryMatchRepository) GetByID(id string) (*entity.Match, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	match, ok := r.matches[id]
	if !ok {
		return nil, errors.New("match not found")
	}
	return match, nil
}
