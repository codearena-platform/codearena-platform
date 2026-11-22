package resource

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func scopedResourceContext() context.Context {
	return middlewares.AddMetadata(context.Background(), sdk.Metadata{
		User:       &sdk.User{Id: "user-1", ProjectId: "project-a"},
		ProjectIds: []string{"project-a"},
	})
}

func TestResourceScoping_CreateRejectsOtherProject(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)

	err := svc.Create(scopedResourceContext(), &sdk.Resource{ID: "resource-1", ProjectId: "project-b"})

	assert.ErrorIs(t, err, middlewares.ErrCrossTenantAccess)
	store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestResourceScoping_GetByIdDoesNotLeakOtherTenant(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)
	ctx := scopedResourceContext()
	store.On("GetById", ctx, "resource-1").Return(&sdk.Resource{ID: "resource-1", ProjectId: "project-b"}, nil)

	resource, err := svc.GetById(ctx, "resource-1")

	assert.ErrorIs(t, err, sdk.ErrResourceNotFound)
	assert.Nil(t, resource)
}

func TestResourceScoping_DeleteRejectsOtherTenant(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)
	ctx := scopedResourceContext()
	store.On("GetById", ctx, "resource-1").Return(&sdk.Resource{ID: "resource-1", ProjectId: "project-b"}, nil)

	err := svc.Delete(ctx, "resource-1")

	assert.ErrorIs(t, err, sdk.ErrResourceNotFound)
	store.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestResourceScoping_SearchUsesVerifiedScope(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)
	ctx := scopedResourceContext()
	store.On("GetAll", ctx, mock.MatchedBy(func(query sdk.ResourceQuery) bool {
		return len(query.ProjectIds) == 1 && query.ProjectIds[0] == "project-a"
	})).Return(&sdk.ResourceList{}, nil)

	_, err := svc.Search(ctx, sdk.ResourceQuery{ProjectIds: []string{"project-b"}})

	assert.NoError(t, err)
	store.AssertExpectations(t)
}
