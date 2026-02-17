package policy

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/repository"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

type service struct {
	repo repository.PolicyRepository
}

func NewService(repo repository.PolicyRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetAll(ctx context.Context, query sdk.PolicyQuery) (*sdk.PolicyList, error) {
	return s.repo.GetAll(ctx, query)
}
