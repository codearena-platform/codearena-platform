package authprovider

import (
	"context"
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/repository"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/authprovider/github"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/authprovider/google"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/authprovider/microsoft"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/authprovider/oidc"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/project"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type service struct {
	repo repository.AuthProviderRepository
	p    project.Service
}

func NewService(repo repository.AuthProviderRepository, p project.Service) Service {
	return &service{
		repo: repo,
		p:    p,
	}
}

func (s service) GetAll(ctx context.Context, params sdk.AuthProviderQueryParams) ([]sdk.AuthProvider, error) {
	projectIds := params.ProjectIds
	if len(projectIds) == 0 {
		projectIds = middlewares.GetProjects(ctx)
	}
	return s.repo.GetAll(ctx, projectIds)
}

func (s service) Get(ctx context.Context, id string, dontCheckProjects bool) (*sdk.AuthProvider, error) {
	ap, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if dontCheckProjects {
		return ap, nil
	}

	// check if the project exists
	projectIdsMap := utils.Reduce(middlewares.GetProjects(ctx), func(ini map[string]bool, p string) map[string]bool { ini[p] = true; return ini }, map[string]bool{})
	if _, ok := projectIdsMap[ap.ProjectId]; !ok {
		return nil, sdk.ErrAuthProviderNotFound
	}
	return ap, nil
}
func (s service) Create(ctx context.Context, provider *sdk.AuthProvider) error {
	// check if the project exists
	projectIdsMap := utils.Reduce(middlewares.GetProjects(ctx), func(ini map[string]bool, p string) map[string]bool { ini[p] = true; return ini }, map[string]bool{})
	if _, ok := projectIdsMap[provider.ProjectId]; !ok {
		return sdk.ErrProjectNotFound
	}
	return s.repo.Create(ctx, provider)
}
func (s service) Update(ctx context.Context, provider *sdk.AuthProvider) error {
	// check if the project exists
	projectIdsMap := utils.Reduce(middlewares.GetProjects(ctx), func(ini map[string]bool, p string) map[string]bool { ini[p] = true; return ini }, map[string]bool{})
	if _, ok := projectIdsMap[provider.ProjectId]; !ok {
		return sdk.ErrProjectNotFound
	}
	return s.repo.Update(ctx, provider)
}

func (s service) GetProvider(ctx context.Context, v sdk.AuthProvider) (sdk.ServiceProvider, error) {
	switch v.Provider {
	case sdk.AuthProviderTypeGoogle:
		return google.NewAuthProvider(v), nil
	case sdk.AuthProviderTypeMicrosoft:
		return microsoft.NewAuthProvider(v), nil
	case sdk.AuthProviderTypeGitHub:
		return github.NewAuthProvider(v), nil
	case sdk.AuthProviderTypeOIDC:
		return oidc.NewAuthProvider(v), nil
	default:
		return nil, fmt.Errorf("unknown auth provider: %s", v.Provider)
	}
}
