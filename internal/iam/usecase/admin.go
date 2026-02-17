package usecase

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/client"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/project"
)

type AdminUseCase interface {
	// Project operations
	CreateProject(ctx context.Context, project *entity.Project) error
	UpdateProject(ctx context.Context, project *entity.Project) error
	GetProject(ctx context.Context, id string) (*entity.Project, error)
	GetProjectByName(ctx context.Context, name string) (*entity.Project, error)
	GetAllProjects(ctx context.Context) ([]entity.Project, error)

	// Client operations
	CreateClient(ctx context.Context, client *entity.Client) error
	UpdateClient(ctx context.Context, client *entity.Client) error
	GetClient(ctx context.Context, id string, dontCheckProjects bool) (*entity.Client, error)
	GetAllClients(ctx context.Context, queryParams sdk.ClientQueryParams) ([]entity.Client, error)
	GetIAMClients(ctx context.Context, params sdk.ClientQueryParams) ([]entity.Client, error)
	RegenerateClientSecret(ctx context.Context, clientId string) (*entity.Client, error)
	VerifyClientSecret(plainSecret, hashedSecret string) error
}

type adminUseCase struct {
	projectSvc project.Service
	clientSvc  client.Service
}

func NewAdminUseCase(projectSvc project.Service, clientSvc client.Service) AdminUseCase {
	return &adminUseCase{
		projectSvc: projectSvc,
		clientSvc:  clientSvc,
	}
}

func (u *adminUseCase) CreateProject(ctx context.Context, project *entity.Project) error {
	return u.projectSvc.Create(ctx, project)
}

func (u *adminUseCase) UpdateProject(ctx context.Context, project *entity.Project) error {
	return u.projectSvc.Update(ctx, project)
}

func (u *adminUseCase) GetProject(ctx context.Context, id string) (*entity.Project, error) {
	return u.projectSvc.GetById(ctx, id)
}

func (u *adminUseCase) GetProjectByName(ctx context.Context, name string) (*entity.Project, error) {
	return u.projectSvc.GetByName(ctx, name)
}

func (u *adminUseCase) GetAllProjects(ctx context.Context) ([]entity.Project, error) {
	return u.projectSvc.GetAll(ctx)
}

func (u *adminUseCase) CreateClient(ctx context.Context, client *entity.Client) error {
	return u.clientSvc.Create(ctx, client)
}

func (u *adminUseCase) UpdateClient(ctx context.Context, client *entity.Client) error {
	return u.clientSvc.Update(ctx, client)
}

func (u *adminUseCase) GetClient(ctx context.Context, id string, dontCheckProjects bool) (*entity.Client, error) {
	return u.clientSvc.Get(ctx, id, dontCheckProjects)
}

func (u *adminUseCase) GetAllClients(ctx context.Context, queryParams sdk.ClientQueryParams) ([]entity.Client, error) {
	return u.clientSvc.GetAll(ctx, queryParams)
}

func (u *adminUseCase) GetIAMClients(ctx context.Context, params sdk.ClientQueryParams) ([]entity.Client, error) {
	return u.clientSvc.GetIAMClients(ctx, params)
}

func (u *adminUseCase) RegenerateClientSecret(ctx context.Context, clientId string) (*entity.Client, error) {
	return u.clientSvc.RegenerateSecret(ctx, clientId)
}

func (u *adminUseCase) VerifyClientSecret(plainSecret, hashedSecret string) error {
	return u.clientSvc.VerifySecret(plainSecret, hashedSecret)
}
