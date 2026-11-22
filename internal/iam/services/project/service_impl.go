package project

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
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
	projects, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if middlewares.GetUser(ctx) == nil {
		return projects, nil
	}
	allowed := map[string]bool{}
	for _, id := range middlewares.GetProjects(ctx) {
		allowed[id] = true
	}
	scoped := make([]sdk.Project, 0, len(projects))
	for _, project := range projects {
		if allowed[project.Id] {
			scoped = append(scoped, project)
		}
	}
	return scoped, nil
}

func (s service) GetByName(ctx context.Context, name string) (*sdk.Project, error) {
	if len(name) == 0 {
		return nil, sdk.ErrProjectNotFound
	}
	project, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	// Do not reveal whether a project exists outside the caller's scope.
	if err := middlewares.AuthorizeProject(ctx, project.Id); err != nil {
		return nil, sdk.ErrProjectNotFound
	}
	return project, nil
}

func (s service) GetById(ctx context.Context, id string) (*sdk.Project, error) {
	if len(id) == 0 {
		return nil, sdk.ErrProjectNotFound
	}
	project, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	// Do not reveal whether a project exists outside the caller's scope.
	if err := middlewares.AuthorizeProject(ctx, project.Id); err != nil {
		return nil, sdk.ErrProjectNotFound
	}
	return project, nil
}

func (s service) Create(ctx context.Context, project *sdk.Project) error {
	if err := middlewares.RequireProjectScope(ctx); err != nil {
		return err
	}
	return s.repo.Create(ctx, project)
}

func (s service) Update(ctx context.Context, project *sdk.Project) error {
	if err := middlewares.AuthorizeProject(ctx, project.Id); err != nil {
		return err
	}
	return s.repo.Update(ctx, project)
}
