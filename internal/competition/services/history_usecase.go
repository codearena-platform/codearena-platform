package services

import (
	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
)

type HistoryUsecase struct {
	matchRepo repository.MatchRepository
	botRepo   repository.BotRepository
}

func NewHistoryUsecase(mr repository.MatchRepository, br repository.BotRepository) *HistoryUsecase {
	return &HistoryUsecase{
		matchRepo: mr,
		botRepo:   br,
	}
}

func (u *HistoryUsecase) GetGlobalHistory(limit int) ([]*entity.Match, error) {
	// For v1, we retrieve from MatchRepository.
	// We might need to add ListFinished to MatchRepository interface if not present.
	return u.matchRepo.ListFinished(limit)
}

func (u *HistoryUsecase) GetBotHistory(botID string) ([]*entity.Match, error) {
	return u.matchRepo.GetByBotID(botID)
}

func (u *HistoryUsecase) GetTournamentHistory(tournamentID string) ([]*entity.Match, error) {
	return u.matchRepo.GetByTournamentID(tournamentID)
}
