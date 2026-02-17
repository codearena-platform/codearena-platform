package memory

import (
	"sync"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
)

type InMemoryBotRepository struct {
	mu   sync.RWMutex
	bots map[string]*entity.Bot
}

func NewInMemoryBotRepository() *InMemoryBotRepository {
	return &InMemoryBotRepository{
		bots: make(map[string]*entity.Bot),
	}
}

func (r *InMemoryBotRepository) Save(bot *entity.Bot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bots[bot.ID] = bot
	return nil
}

func (r *InMemoryBotRepository) GetByID(id string) (*entity.Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bot, ok := r.bots[id]
	if !ok {
		// Auto-register mock for now if not found (dev mode)
		return &entity.Bot{
			ID:     id,
			Name:   "Mock Bot " + id,
			Rating: 1200.0, // Default ELO
		}, nil
		// return nil, errors.New("bot not found")
	}
	return bot, nil
}

func (r *InMemoryBotRepository) GetByUserID(userID string) ([]*entity.Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entity.Bot
	for _, b := range r.bots {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (r *InMemoryBotRepository) UpdateRating(botID string, newRating float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	bot, ok := r.bots[botID]
	if !ok {
		// Create if not exists for dev
		bot = &entity.Bot{ID: botID, Rating: 1200.0}
		r.bots[botID] = bot
	}
	bot.Rating = newRating
	return nil
}
