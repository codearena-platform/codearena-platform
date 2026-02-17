package resource

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/repository"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type service struct {
	repo repository.ResourceRepository
	e    utils.Emitter[utils.Event[sdk.Resource], sdk.Resource]
}

func NewService(repo repository.ResourceRepository) Service {
	return service{repo: repo,
		e: utils.NewEmitter[utils.Event[sdk.Resource]]()}
}

func (s service) Search(ctx context.Context, query sdk.ResourceQuery) (*sdk.ResourceList, error) {
	query.ProjectIds = middlewares.GetProjects(ctx)
	return s.repo.GetAll(ctx, query)
}

func (s service) GetById(ctx context.Context, id string) (*sdk.Resource, error) {
	return s.repo.GetById(ctx, id)
}

func (s service) Create(ctx context.Context, resource *sdk.Resource) error {
	_, err := s.repo.Create(ctx, resource)
	if err != nil {
		return err
	}
	s.Emit(newEvent(ctx, constants.EventResourceCreated, *resource, middlewares.GetMetadata(ctx)))
	return nil
}

func (s service) Update(ctx context.Context, resource *sdk.Resource) error {
	return s.repo.Update(ctx, resource)
}

func (s service) Delete(ctx context.Context, id string) error {
	vl, err := s.repo.GetById(ctx, id)
	if err != nil {
		return err
	}
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	s.Emit(newEvent(ctx, constants.EventResourceDeleted, *vl, middlewares.GetMetadata(ctx)))
	return nil
}

func (s service) Emit(event utils.Event[sdk.Resource]) {
	if event == nil {
		return
	}
	s.e.Emit(event)
}

func (s service) Subscribe(eventName constants.Event, subscriber utils.Subscriber[utils.Event[sdk.Resource], sdk.Resource]) {
	s.e.Subscribe(eventName, subscriber)
}

type event struct {
	name     constants.Event
	payload  sdk.Resource
	metadata sdk.Metadata
	ctx      context.Context
}

func (e event) Name() constants.Event {
	return e.name
}

func (e event) Payload() sdk.Resource {
	return e.payload
}

func (e event) Metadata() sdk.Metadata {
	return e.metadata
}

func (e event) Context() context.Context {
	return e.ctx
}

func newEvent(ctx context.Context, name constants.Event, payload sdk.Resource, metadata sdk.Metadata) utils.Event[sdk.Resource] {
	return event{ctx: ctx, name: name, payload: payload, metadata: metadata}
}
