package services

import (
	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
)

type ScoreboardService struct {
	tournamentRepo repository.TournamentRepository
	leagueRepo     repository.LeagueRepository
	botRepo        repository.BotRepository
}

func NewScoreboardService(tr repository.TournamentRepository, lr repository.LeagueRepository, br repository.BotRepository) *ScoreboardService {
	return &ScoreboardService{
		tournamentRepo: tr,
		leagueRepo:     lr,
		botRepo:        br,
	}
}

func (s *ScoreboardService) GetTournamentBracket(tournamentID string) ([]*entity.Bracket, error) {
	return s.tournamentRepo.GetBrackets(tournamentID)
}

func (s *ScoreboardService) GetDivisionStandings(divisionID string) ([]*entity.Standing, error) {
	return s.leagueRepo.GetStandings(divisionID)
}

func (s *ScoreboardService) GetGlobalHallOfFame() ([]*entity.Bot, error) {
	// Implementation: top 10 by Elo
	return s.botRepo.GetTop(10)
}
