package services

import (
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/codearena-platform/codearena-platform/internal/competition/services/repository"
)

type HubUsecase struct {
	hubRepo        repository.HubRepository
	tournamentRepo repository.TournamentRepository
}

func NewHubUsecase(hr repository.HubRepository, tr repository.TournamentRepository) *HubUsecase {
	return &HubUsecase{
		hubRepo:        hr,
		tournamentRepo: tr,
	}
}

func (u *HubUsecase) CreateHub(name, ownerID, description string, hubType entity.HubType, visibility entity.HubVisibility) (*entity.Hub, error) {
	hub := &entity.Hub{
		ID:          fmt.Sprintf("hub_%d", time.Now().UnixNano()),
		Name:        name,
		Type:        hubType,
		Visibility:  visibility,
		OwnerID:     ownerID,
		Description: description,
		CreatedAt:   time.Now().Unix(),
	}

	if err := u.hubRepo.Save(hub); err != nil {
		return nil, err
	}

	// Add owner as a member
	membership := &entity.HubMembership{
		HubID:    hub.ID,
		UserID:   ownerID,
		Role:     "OWNER",
		JoinedAt: hub.CreatedAt,
	}
	u.hubRepo.SaveMembership(membership)

	return hub, nil
}

func (u *HubUsecase) AddMember(hubID, userID, role string) error {
	membership := &entity.HubMembership{
		HubID:    hubID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now().Unix(),
	}
	return u.hubRepo.SaveMembership(membership)
}

func (u *HubUsecase) ListUserHubs(userID string) ([]*entity.Hub, error) {
	return u.hubRepo.ListByUser(userID)
}

func (u *HubUsecase) GetHubTournaments(hubID string) ([]*entity.Tournament, error) {
	return u.tournamentRepo.ListByHub(hubID)
}
