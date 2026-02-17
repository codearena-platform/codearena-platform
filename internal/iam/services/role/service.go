package role

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type Service interface {
	Create(ctx context.Context, role *sdk.Role) error
	Update(ctx context.Context, role *sdk.Role) error
	GetById(ctx context.Context, id string) (*sdk.Role, error)
	GetAll(ctx context.Context, query sdk.RoleQuery) (*sdk.RoleList, error)
	AddResource(ctx context.Context, roleId string, resource sdk.Resources) error
	RemoveResourceFromAll(ctx context.Context, resourceKey string) error
	utils.Emitter[utils.Event[sdk.Role], sdk.Role]
}
