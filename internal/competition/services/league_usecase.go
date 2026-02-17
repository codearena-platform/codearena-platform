package services

import (
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
)

type LeagueUsecase struct {
	leagueRepo repository.LeagueRepository
	botRepo    repository.BotRepository
}

func NewLeagueUsecase(lr repository.LeagueRepository, br repository.BotRepository) *LeagueUsecase {
	return &LeagueUsecase{
		leagueRepo: lr,
		botRepo:    br,
	}
}

func (u *LeagueUsecase) CreateLeague(tournamentID, season string) (*entity.League, error) {
	league := &entity.League{
		ID:           fmt.Sprintf("lg_%d", time.Now().UnixNano()),
		TournamentID: tournamentID,
		Season:       season,
	}

	if err := u.leagueRepo.Save(league); err != nil {
		return nil, err
	}

	return league, nil
}

func (u *LeagueUsecase) SetupDivisions(leagueID string) error {
	divisions := []entity.DivisionName{
		entity.DivisionSerieA,
		entity.DivisionSerieB,
		entity.DivisionSerieC,
		entity.DivisionAmateur,
	}

	for _, name := range divisions {
		div := &entity.Division{
			ID:              fmt.Sprintf("div_%s_%d", name, time.Now().UnixNano()),
			LeagueID:        leagueID,
			Name:            name,
			Capacity:        20,
			PromotionSlots:  4,
			RelegationSlots: 4,
		}

		if err := u.leagueRepo.SaveDivision(div); err != nil {
			return err
		}
	}

	return nil
}

func (u *LeagueUsecase) UpdateStanding(divisionID, botID string, result string) error {
	standings, _ := u.leagueRepo.GetStandings(divisionID)
	var current *entity.Standing

	for _, s := range standings {
		if s.BotID == botID {
			current = s
			break
		}
	}

	if current == nil {
		current = &entity.Standing{
			DivisionID: divisionID,
			BotID:      botID,
		}
	}

	current.Matches++
	switch result {
	case "WIN":
		current.Points += 3
		current.Wins++
	case "DRAW":
		current.Points += 1
		current.Draws++
	case "LOSS":
		current.Losses++
	}

	return u.leagueRepo.SaveStanding(current)
}
