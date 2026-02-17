package services

import (
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
)

type TournamentUsecase struct {
	tournamentRepo repository.TournamentRepository
	matchRepo      repository.MatchRepository
}

func NewTournamentUsecase(tr repository.TournamentRepository, mr repository.MatchRepository) *TournamentUsecase {
	return &TournamentUsecase{
		tournamentRepo: tr,
		matchRepo:      mr,
	}
}

func (u *TournamentUsecase) CreateTournament(name string, tType entity.TournamentType) (*entity.Tournament, error) {
	tournament := &entity.Tournament{
		ID:        fmt.Sprintf("t_%d", time.Now().UnixNano()),
		Name:      name,
		Type:      tType,
		Status:    entity.TournamentDraft,
		CreatedAt: time.Now().Unix(),
	}

	if err := u.tournamentRepo.Save(tournament); err != nil {
		return nil, err
	}

	return tournament, nil
}

func (u *TournamentUsecase) GenerateBrackets(tournamentID string, participantIDs []string) error {
	// Simple bracket generation for power of 2 participants
	// level 1: Final, level 2: Semi, level 4: Quarter
	num := len(participantIDs)
	if num < 2 {
		return fmt.Errorf("at least 2 participants required")
	}

	// For MVP, we'll just create the first round matches
	for i := 0; i < num; i += 2 {
		if i+1 >= num {
			break // Odd number of participants handled as bye in full impl
		}

		match := &entity.Match{
			ID:           fmt.Sprintf("tm_%d", time.Now().UnixNano()),
			TournamentID: tournamentID,
			Status:       entity.MatchReady,
			Participants: []string{participantIDs[i], participantIDs[i+1]},
			StartedAt:    time.Now().Unix(),
		}

		if err := u.matchRepo.Save(match); err != nil {
			return err
		}

		bracket := &entity.Bracket{
			ID:           fmt.Sprintf("b_%d", time.Now().UnixNano()),
			TournamentID: tournamentID,
			Level:        num / 2, // Example level logic
			MatchID:      match.ID,
			Position:     i / 2,
		}

		if err := u.tournamentRepo.SaveBracket(bracket); err != nil {
			return err
		}
	}

	return nil
}
