package usecase

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/user"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
)

type IdentityUseCase interface {
	// User operations
	CreateUser(ctx context.Context, user *entity.User) error
	UpdateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string, projectId string) (*entity.User, error)
	GetUserById(ctx context.Context, id string) (*entity.User, error)
	GetUserByPhone(ctx context.Context, phone string, projectId string) (*entity.User, error)
	GetAllUsers(ctx context.Context, query sdk.UserQuery) (*sdk.UserList, error)

	// Administrative user operations
	AddRoleToUser(ctx context.Context, userId, roleId string) error
	RemoveRoleFromUser(ctx context.Context, userId, roleId string) error
	AddResourceToUser(ctx context.Context, userId string, request sdk.AddUserResourceRequest) error
	AddPolicyToUser(ctx context.Context, userId string, policies map[string]sdk.UserPolicy) error
	RemovePolicyFromUser(ctx context.Context, userId string, policyIds []string) error
	TransferOwnership(ctx context.Context, userId, newOwnerId string) error
	CopyUserResources(ctx context.Context, sourceUserId, targetUserId string) error

	// Events
	Subscribe(eventName constants.Event, subscriber utils.Subscriber[utils.Event[entity.User], entity.User])
}

type identityUseCase struct {
	userSvc user.Service
}

func NewIdentityUseCase(userSvc user.Service) IdentityUseCase {
	return &identityUseCase{
		userSvc: userSvc,
	}
}

func (u *identityUseCase) CreateUser(ctx context.Context, user *entity.User) error {
	return u.userSvc.Create(ctx, user)
}

func (u *identityUseCase) UpdateUser(ctx context.Context, user *entity.User) error {
	return u.userSvc.Update(ctx, user)
}

func (u *identityUseCase) GetUserByEmail(ctx context.Context, email string, projectId string) (*entity.User, error) {
	return u.userSvc.GetByEmail(ctx, email, projectId)
}

func (u *identityUseCase) GetUserById(ctx context.Context, id string) (*entity.User, error) {
	return u.userSvc.GetById(ctx, id)
}

func (u *identityUseCase) GetUserByPhone(ctx context.Context, phone string, projectId string) (*entity.User, error) {
	return u.userSvc.GetByPhone(ctx, phone, projectId)
}

func (u *identityUseCase) GetAllUsers(ctx context.Context, query sdk.UserQuery) (*sdk.UserList, error) {
	return u.userSvc.GetAll(ctx, query)
}

func (u *identityUseCase) AddRoleToUser(ctx context.Context, userId, roleId string) error {
	return u.userSvc.AddRoleToUser(ctx, userId, roleId)
}

func (u *identityUseCase) RemoveRoleFromUser(ctx context.Context, userId, roleId string) error {
	return u.userSvc.RemoveRoleFromUser(ctx, userId, roleId)
}

func (u *identityUseCase) AddResourceToUser(ctx context.Context, userId string, request sdk.AddUserResourceRequest) error {
	return u.userSvc.AddResourceToUser(ctx, userId, request)
}

func (u *identityUseCase) AddPolicyToUser(ctx context.Context, userId string, policies map[string]sdk.UserPolicy) error {
	return u.userSvc.AddPolicyToUser(ctx, userId, policies)
}

func (u *identityUseCase) RemovePolicyFromUser(ctx context.Context, userId string, policyIds []string) error {
	return u.userSvc.RemovePolicyFromUser(ctx, userId, policyIds)
}

func (u *identityUseCase) TransferOwnership(ctx context.Context, userId, newOwnerId string) error {
	return u.userSvc.TransferOwnership(ctx, userId, newOwnerId)
}

func (u *identityUseCase) CopyUserResources(ctx context.Context, sourceUserId, targetUserId string) error {
	return u.userSvc.CopyUserResources(ctx, sourceUserId, targetUserId)
}

func (u *identityUseCase) Subscribe(eventName constants.Event, subscriber utils.Subscriber[utils.Event[entity.User], entity.User]) {
	u.userSvc.Subscribe(eventName, subscriber)
}
