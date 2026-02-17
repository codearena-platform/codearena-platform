package authprovider

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

type Store interface {
	GetAll(ctx context.Context, params sdk.AuthProviderQueryParams) ([]sdk.AuthProvider, error)
	Get(ctx context.Context, id string) (*sdk.AuthProvider, error)
	Create(ctx context.Context, provider *sdk.AuthProvider) error
	Update(ctx context.Context, provider *sdk.AuthProvider) error
}
