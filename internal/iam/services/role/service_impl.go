package role

import (
	"context"
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/repository"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type service struct {
	repo repository.RoleRepository
	e    utils.Emitter[utils.Event[sdk.Role], sdk.Role]
}

func NewService(repo repository.RoleRepository) Service {
	return &service{
		repo: repo,
		e:    utils.NewEmitter[utils.Event[sdk.Role]](),
	}
}
func (s *service) Create(ctx context.Context, role *sdk.Role) error {
	return s.repo.Create(ctx, role)
}

func (s *service) Update(ctx context.Context, role *sdk.Role) error {
	err := s.repo.Update(ctx, role)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	s.Emit(newEvent(ctx, constants.EventRoleUpdated, *role, middlewares.GetMetadata(ctx)))
	return nil
}

func (s *service) GetById(ctx context.Context, id string) (*sdk.Role, error) {
	return s.repo.GetById(ctx, id)
}

func (s *service) GetAll(ctx context.Context, query sdk.RoleQuery) (*sdk.RoleList, error) {
	query.ProjectIds = middlewares.GetProjects(ctx)
	return s.repo.GetAll(ctx, query)
}

func (s *service) AddResource(ctx context.Context, roleId string, resource sdk.Resources) error {
	role, err := s.GetById(ctx, roleId)
	if err != nil {
		return err
	}
	if len(role.Resources) == 0 {
		role.Resources = map[string]sdk.Resources{}
	}
	role.Resources[resource.Key] = resource
	return s.Update(ctx, role)
}

func (s *service) RemoveResourceFromAll(ctx context.Context, resourceKey string) error {
	return s.repo.RemoveResourceFromAll(ctx, resourceKey)
}

func (s service) Emit(event utils.Event[sdk.Role]) {
	if event == nil {
		return
	}
	s.e.Emit(event)
}

func (s service) Subscribe(eventName constants.Event, subscriber utils.Subscriber[utils.Event[sdk.Role], sdk.Role]) {
	s.e.Subscribe(eventName, subscriber)
}

type event struct {
	name     constants.Event
	payload  sdk.Role
	metadata sdk.Metadata
	ctx      context.Context
}

func (e event) Name() constants.Event {
	return e.name
}

func (e event) Payload() sdk.Role {
	return e.payload
}

func (e event) Metadata() sdk.Metadata {
	return e.metadata
}

func (e event) Context() context.Context {
	return e.ctx
}

func newEvent(ctx context.Context, name constants.Event, payload sdk.Role, metadata sdk.Metadata) utils.Event[sdk.Role] {
	return event{ctx: ctx, name: name, payload: payload, metadata: metadata}
}
