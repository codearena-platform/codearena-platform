package project

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/repository"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

type service struct {
	repo repository.ProjectRepository
}

func NewService(repo repository.ProjectRepository) Service {
	return service{repo: repo}
}

func (s service) GetAll(ctx context.Context) ([]sdk.Project, error) {
	return s.repo.GetAll(ctx)
}

func (s service) GetByName(ctx context.Context, name string) (*sdk.Project, error) {
	if len(name) == 0 {
		return nil, sdk.ErrProjectNotFound
	}
	return s.repo.GetByName(ctx, name)
}

func (s service) GetById(ctx context.Context, id string) (*sdk.Project, error) {
	if len(id) == 0 {
		return nil, sdk.ErrProjectNotFound
	}
	return s.repo.GetById(ctx, id)
}

func (s service) Create(ctx context.Context, project *sdk.Project) error {
	return s.repo.Create(ctx, project)
}

func (s service) Update(ctx context.Context, project *sdk.Project) error {
	return s.repo.Update(ctx, project)
}
