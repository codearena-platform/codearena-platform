package client

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type Service interface {
	GetAll(ctx context.Context, queryParams sdk.ClientQueryParams) ([]sdk.Client, error)
	GetIAMClients(ctx context.Context, params sdk.ClientQueryParams) ([]sdk.Client, error)
	Get(ctx context.Context, id string, dontCheckProjects bool) (*sdk.Client, error)
	Create(ctx context.Context, client *sdk.Client) error
	Update(ctx context.Context, client *sdk.Client) error
	VerifySecret(plainSecret, hashedSecret string) error
	RegenerateSecret(ctx context.Context, clientId string) (*sdk.Client, error)
	utils.Emitter[utils.Event[sdk.Client], sdk.Client]
}
