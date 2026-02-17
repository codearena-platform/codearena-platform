package repository

import (
	"context"
	"strings"

	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
)

type PolicyRepository interface {
	GetAll(ctx context.Context, query sdk.PolicyQuery) (*sdk.PolicyList, error)
}

type policyRepository struct{}

func NewPolicyRepository() PolicyRepository {
	return &policyRepository{}
}

func (r *policyRepository) GetAll(ctx context.Context, query sdk.PolicyQuery) (*sdk.PolicyList, error) {
	// Static policies hardcoded to avoid import cycle with system services
	policies := []entity.Policy{
		{
			Id:          "@policy/system/access_to_created_resource",
			Name:        "User get access to the resource created by the user",
			Description: "This policy grants the user access to the resource created by them.",
			Definition: entity.PolicyDefinition{
				Arguments: []entity.PolicyArgument{
					{
						Name:        "@userId",
						Description: "The user to whom the resource access is granted.",
						DataType:    constants.User,
					},
				},
			},
		},
		{
			Id:          "@policy/system/add_resources_to_role",
			Name:        "Add the resources created by a user to role specified in user policy",
			Description: "This policy adds the created resource to the role specified in the user policy.",
			Definition: entity.PolicyDefinition{
				Arguments: []entity.PolicyArgument{
					{
						Name:        "@roleId",
						Description: "The role to which the resource access is granted.",
						DataType:    constants.Role,
					},
				},
			},
		},
		{
			Id:          "@policy/system/add_resources_to_user",
			Name:        "Add the resources created by a user to another user specified in user policy",
			Description: "This policy adds the created resource to the user specified in the user policy.",
			Definition: entity.PolicyDefinition{
				Arguments: []entity.PolicyArgument{
					{
						Name:        "@userId",
						Description: "The user to whom the resource access is granted.",
						DataType:    constants.User,
					},
				},
			},
		},
	}

	limit := query.Limit
	skip := query.Skip
	if limit <= 0 {
		limit = 10
	}
	if skip < 0 {
		skip = 0
	}

	var pls []entity.Policy
	for _, policy := range policies {
		if query.Query == "" || (query.Query != "" && strings.Contains(strings.ToLower(policy.Name), strings.ToLower(query.Query))) {
			pls = append(pls, policy)
		}
	}

	total := len(pls)
	if skip >= int64(total) {
		return &sdk.PolicyList{
			Policies: []entity.Policy{},
			Total:    total,
			Skip:     skip,
			Limit:    limit,
		}, nil
	}

	end := skip + limit
	if end > int64(total) {
		end = int64(total)
	}

	result := pls[skip:end]

	return &sdk.PolicyList{
		Policies: result,
		Total:    total,
		Skip:     skip,
		Limit:    limit,
	}, nil
}
