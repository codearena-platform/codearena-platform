package role

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func scopedRoleContext() context.Context {
	return middlewares.AddMetadata(context.Background(), sdk.Metadata{
		User:       &sdk.User{Id: "user-1", ProjectId: "project-a"},
		ProjectIds: []string{"project-a"},
	})
}

func TestRoleScoping_CreateRejectsOtherProject(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)

	err := svc.Create(scopedRoleContext(), &sdk.Role{Id: "role-1", ProjectId: "project-b"})

	assert.ErrorIs(t, err, middlewares.ErrCrossTenantAccess)
	store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRoleScoping_GetByIdDoesNotLeakOtherTenant(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)
	ctx := scopedRoleContext()
	store.On("GetById", ctx, "role-1").Return(&sdk.Role{Id: "role-1", ProjectId: "project-b"}, nil)

	role, err := svc.GetById(ctx, "role-1")

	assert.ErrorIs(t, err, sdk.ErrRoleNotFound)
	assert.Nil(t, role)
}

func TestRoleScoping_GetAllUsesVerifiedScope(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)
	ctx := scopedRoleContext()
	store.On("GetAll", ctx, mock.MatchedBy(func(query sdk.RoleQuery) bool {
		return len(query.ProjectIds) == 1 && query.ProjectIds[0] == "project-a"
	})).Return(&sdk.RoleList{}, nil)

	_, err := svc.GetAll(ctx, sdk.RoleQuery{ProjectIds: []string{"project-b"}})

	assert.NoError(t, err)
	store.AssertExpectations(t)
}
