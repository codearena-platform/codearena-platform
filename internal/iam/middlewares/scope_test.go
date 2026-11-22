package middlewares

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/stretchr/testify/assert"
)

func scopedContext(projectIDs ...string) context.Context {
	return AddMetadata(context.Background(), sdk.Metadata{
		User:       &sdk.User{Id: "user-1", ProjectId: projectIDs[0]},
		ProjectIds: projectIDs,
	})
}

func TestRequireProjectScope(t *testing.T) {
	assert.NoError(t, RequireProjectScope(context.Background()), "internal calls have no authenticated user")

	noScope := AddMetadata(context.Background(), sdk.Metadata{User: &sdk.User{Id: "user-1"}})
	assert.ErrorIs(t, RequireProjectScope(noScope), ErrProjectScopeMissing)

	assert.NoError(t, RequireProjectScope(scopedContext("project-a")))
}

func TestAuthorizeProject(t *testing.T) {
	assert.NoError(t, AuthorizeProject(context.Background(), "project-b"), "internal calls may cross projects")

	ctx := scopedContext("project-a")
	assert.NoError(t, AuthorizeProject(ctx, "project-a"))
	assert.ErrorIs(t, AuthorizeProject(ctx, "project-b"), ErrCrossTenantAccess)
	assert.NoError(t, AuthorizeProject(ctx, ""), "legacy unscoped documents are not cross-tenant")

	noScope := AddMetadata(context.Background(), sdk.Metadata{User: &sdk.User{Id: "user-1"}})
	assert.ErrorIs(t, AuthorizeProject(noScope, "project-a"), ErrProjectScopeMissing)
}
