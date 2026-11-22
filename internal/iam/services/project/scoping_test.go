package project

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func scopedProjectContext() context.Context {
	return middlewares.AddMetadata(context.Background(), sdk.Metadata{
		User:       &sdk.User{Id: "user-1", ProjectId: "project-a"},
		ProjectIds: []string{"project-a"},
	})
}

func TestProjectScoping_GetByIdDoesNotLeakOtherTenant(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)
	ctx := scopedProjectContext()
	store.On("GetById", ctx, "project-b").Return(&sdk.Project{Id: "project-b"}, nil)

	project, err := svc.GetById(ctx, "project-b")

	assert.ErrorIs(t, err, sdk.ErrProjectNotFound)
	assert.Nil(t, project)
}

func TestProjectScoping_UpdateRejectsOtherTenant(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)

	err := svc.Update(scopedProjectContext(), &sdk.Project{Id: "project-b"})

	assert.ErrorIs(t, err, middlewares.ErrCrossTenantAccess)
	store.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestProjectScoping_GetAllFiltersToVerifiedScope(t *testing.T) {
	store := &MockStore{}
	svc := NewService(store)
	ctx := scopedProjectContext()
	store.On("GetAll", ctx).Return([]sdk.Project{{Id: "project-a"}, {Id: "project-b"}}, nil)

	projects, err := svc.GetAll(ctx)

	assert.NoError(t, err)
	assert.Equal(t, []sdk.Project{{Id: "project-a"}}, projects)
}
