package user

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func scopedUserContext() context.Context {
	return middlewares.AddMetadata(context.Background(), sdk.Metadata{
		User:       &sdk.User{Id: "user-1", ProjectId: "project-a"},
		ProjectIds: []string{"project-a"},
	})
}

func TestUserScoping_CreateRejectsOtherProject(t *testing.T) {
	svc, store, _ := setupUserService()
	other := createTestUser()
	other.ProjectId = "project-b"

	err := svc.Create(scopedUserContext(), other)

	assert.ErrorIs(t, err, middlewares.ErrCrossTenantAccess)
	store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserScoping_GetByIdDoesNotLeakOtherTenant(t *testing.T) {
	svc, store, _ := setupUserService()
	ctx := scopedUserContext()
	other := createTestUser()
	other.ProjectId = "project-b"
	store.On("GetById", ctx, "user-2").Return(other, nil)

	user, err := svc.GetById(ctx, "user-2")

	assert.ErrorIs(t, err, sdk.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestUserScoping_MutationOfOtherTenantIsRejected(t *testing.T) {
	svc, store, _ := setupUserService()
	ctx := scopedUserContext()
	other := createTestUser()
	other.ProjectId = "project-b"
	store.On("GetById", ctx, "user-2").Return(other, nil)

	err := svc.AddPolicyToUser(ctx, "user-2", map[string]sdk.UserPolicy{})

	assert.ErrorIs(t, err, sdk.ErrUserNotFound)
	store.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUserScoping_GetAllUsesVerifiedScope(t *testing.T) {
	svc, store, _ := setupUserService()
	ctx := scopedUserContext()
	store.On("GetAll", ctx, mock.MatchedBy(func(query sdk.UserQuery) bool {
		return len(query.ProjectIds) == 1 && query.ProjectIds[0] == "project-a"
	})).Return(&sdk.UserList{}, nil)

	_, err := svc.GetAll(ctx, sdk.UserQuery{ProjectIds: []string{"project-b"}})

	assert.NoError(t, err)
	store.AssertExpectations(t)
}
