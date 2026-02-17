package services

import (
	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
)

type AnalyticsUsecase struct {
	matchRepo repository.MatchRepository
	botRepo   repository.BotRepository
}

func NewAnalyticsUsecase(mr repository.MatchRepository, br repository.BotRepository) *AnalyticsUsecase {
	return &AnalyticsUsecase{
		matchRepo: mr,
		botRepo:   br,
	}
}

func (u *AnalyticsUsecase) GetGlobalStats() (*entity.GlobalStats, error) {
	rawStats, err := u.matchRepo.GetAggregateStats(nil)
	if err != nil {
		return nil, err
	}

	totalMatches, _ := rawStats["total_matches"].(int32)

	stats := &entity.GlobalStats{
		TotalMatches: int(totalMatches),
		Classes:      make(map[entity.ClassName]*entity.ClassStats),
	}

	// Default empty stats for now
	stats.Classes[entity.ClassTank] = &entity.ClassStats{Class: entity.ClassTank}
	stats.Classes[entity.ClassScout] = &entity.ClassStats{Class: entity.ClassScout}
	stats.Classes[entity.ClassSniper] = &entity.ClassStats{Class: entity.ClassSniper}

	return stats, nil
}

func (u *AnalyticsUsecase) GetBotAnalytics(botID string) (*entity.BotAnalytics, error) {
	matches, err := u.matchRepo.GetByBotID(botID)
	if err != nil {
		return nil, err
	}

	analytics := &entity.BotAnalytics{
		BotID:        botID,
		TotalMatches: len(matches),
	}

	var totalDamage float64
	var totalSurvival float64
	var wins int

	for _, m := range matches {
		if m.WinnerID == botID {
			wins++
		}
		// Extract metrics from metadata if available
		if meta, ok := m.Metadata[botID].(map[string]any); ok {
			if d, ok := meta["damage"].(float64); ok {
				totalDamage += d
			}
			if s, ok := meta["survival_ticks"].(float64); ok {
				totalSurvival += s
			}
		}
	}

	if analytics.TotalMatches > 0 {
		analytics.WinCount = wins
		analytics.WinRate = float64(wins) / float64(analytics.TotalMatches)
		analytics.ADR = totalDamage / float64(analytics.TotalMatches)
		analytics.AvgSurvivalTime = totalSurvival / float64(analytics.TotalMatches)
	}

	return analytics, nil
}
