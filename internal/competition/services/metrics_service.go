package services

import (
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
)

type MetricsService struct {
	matchRepo repository.MatchRepository
}

func NewMetricsService(mr repository.MatchRepository) *MetricsService {
	return &MetricsService{
		matchRepo: mr,
	}
}

func (s *MetricsService) AggregateMatchMetrics(matchID string, events []map[string]any) error {
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return err
	}

	if match.Metadata == nil {
		match.Metadata = make(map[string]any)
	}

	// Example aggregation logic
	var totalDamage float64
	var maxLifeTime int64

	for _, event := range events {
		if dmg, ok := event["damage"].(float64); ok {
			totalDamage += dmg
		}
		if lt, ok := event["lifetime"].(int64); ok {
			if lt > maxLifeTime {
				maxLifeTime = lt
			}
		}
	}

	match.Metadata["total_damage"] = totalDamage
	match.Metadata["max_lifetime"] = maxLifeTime

	return s.matchRepo.Save(match)
}
