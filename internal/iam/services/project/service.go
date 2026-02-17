package project

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

type Service interface {
	GetAll(ctx context.Context) ([]sdk.Project, error)
	GetById(ctx context.Context, id string) (*sdk.Project, error)
	Create(ctx context.Context, project *sdk.Project) error
	Update(ctx context.Context, project *sdk.Project) error
	GetByName(ctx context.Context, name string) (*sdk.Project, error)
}
