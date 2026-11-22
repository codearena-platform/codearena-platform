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

// GetHub returns a hub so callers can enforce owner-only operations.
func (u *HubUsecase) GetHub(id string) (*entity.Hub, error) {
	return u.hubRepo.GetByID(id)
}

// IsMember reports whether the user belongs to the hub.
func (u *HubUsecase) IsMember(hubID, userID string) (bool, error) {
	memberships, err := u.hubRepo.GetMemberships(hubID)
	if err != nil {
		return false, err
	}
	for _, m := range memberships {
		if m.UserID == userID {
			return true, nil
		}
	}
	return false, nil
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
