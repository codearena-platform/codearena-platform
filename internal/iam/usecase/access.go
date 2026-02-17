package usecase

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/policy"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/resource"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/role"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type AccessUseCase interface {
	// Role operations
	CreateRole(ctx context.Context, role *entity.Role) error
	UpdateRole(ctx context.Context, role *entity.Role) error
	GetRole(ctx context.Context, id string) (*entity.Role, error)
	GetAllRoles(ctx context.Context, query sdk.RoleQuery) (*sdk.RoleList, error)

	// Resource operations
	CreateResource(ctx context.Context, resource *entity.Resource) error
	UpdateResource(ctx context.Context, resource *entity.Resource) error
	GetResource(ctx context.Context, id string) (*entity.Resource, error)
	GetAllResources(ctx context.Context, query sdk.ResourceQuery) (*sdk.ResourceList, error)
	DeleteResource(ctx context.Context, id string) error

	// Policy operations
	GetAllPolicies(ctx context.Context, query sdk.PolicyQuery) (*sdk.PolicyList, error)

	// Subscription
	SubscribeRole(eventName constants.Event, subscriber utils.Subscriber[utils.Event[entity.Role], entity.Role])
	SubscribeResource(eventName constants.Event, subscriber utils.Subscriber[utils.Event[entity.Resource], entity.Resource])
}

type accessUseCase struct {
	roleSvc     role.Service
	resourceSvc resource.Service
	policySvc   policy.Service
}

func NewAccessUseCase(roleSvc role.Service, resourceSvc resource.Service, policySvc policy.Service) AccessUseCase {
	return &accessUseCase{
		roleSvc:     roleSvc,
		resourceSvc: resourceSvc,
		policySvc:   policySvc,
	}
}

func (u *accessUseCase) CreateRole(ctx context.Context, role *entity.Role) error {
	return u.roleSvc.Create(ctx, role)
}

func (u *accessUseCase) UpdateRole(ctx context.Context, role *entity.Role) error {
	return u.roleSvc.Update(ctx, role)
}

func (u *accessUseCase) GetRole(ctx context.Context, id string) (*entity.Role, error) {
	return u.roleSvc.GetById(ctx, id)
}

func (u *accessUseCase) GetAllRoles(ctx context.Context, query sdk.RoleQuery) (*sdk.RoleList, error) {
	return u.roleSvc.GetAll(ctx, query)
}

func (u *accessUseCase) CreateResource(ctx context.Context, resource *entity.Resource) error {
	return u.resourceSvc.Create(ctx, resource)
}

func (u *accessUseCase) UpdateResource(ctx context.Context, resource *entity.Resource) error {
	return u.resourceSvc.Update(ctx, resource)
}

func (u *accessUseCase) GetResource(ctx context.Context, id string) (*entity.Resource, error) {
	return u.resourceSvc.GetById(ctx, id)
}

func (u *accessUseCase) GetAllResources(ctx context.Context, query sdk.ResourceQuery) (*sdk.ResourceList, error) {
	return u.resourceSvc.Search(ctx, query)
}

func (u *accessUseCase) DeleteResource(ctx context.Context, id string) error {
	return u.resourceSvc.Delete(ctx, id)
}

func (u *accessUseCase) GetAllPolicies(ctx context.Context, query sdk.PolicyQuery) (*sdk.PolicyList, error) {
	return u.policySvc.GetAll(ctx, query)
}

func (u *accessUseCase) SubscribeRole(eventName constants.Event, subscriber utils.Subscriber[utils.Event[entity.Role], entity.Role]) {
	u.roleSvc.Subscribe(eventName, subscriber)
}

func (u *accessUseCase) SubscribeResource(eventName constants.Event, subscriber utils.Subscriber[utils.Event[entity.Resource], entity.Resource]) {
	u.resourceSvc.Subscribe(eventName, subscriber)
}
