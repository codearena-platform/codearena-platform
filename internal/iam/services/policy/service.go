package policy

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

type Service interface {
	GetAll(ctx context.Context, query sdk.PolicyQuery) (*sdk.PolicyList, error)
}
