package repository

import (
	"context"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func scopedQueryContext() context.Context {
	return middlewares.AddMetadata(context.Background(), sdk.Metadata{
		User:       &sdk.User{Id: "user-1", ProjectId: "project-a"},
		ProjectIds: []string{"project-a"},
	})
}

func bsonField(document bson.D, key string) interface{} {
	for _, element := range document {
		if element.Key == key {
			return element.Value
		}
	}
	return nil
}

func TestProjectIDsForQuery_AuthenticatedCallerCannotExpandScope(t *testing.T) {
	assert.Equal(t, []string{"project-a"}, projectIDsForQuery(scopedQueryContext(), []string{"project-b"}))
	assert.Equal(t, []string{"project-b"}, projectIDsForQuery(context.Background(), []string{"project-b"}))
}

func TestWithProjectScope_AuthenticatedCallerIsScoped(t *testing.T) {
	filter := withProjectScope(scopedQueryContext(), "project_id", bson.D{{Key: "key", Value: "value"}})
	require.Len(t, filter, 2)
	assert.Equal(t, bson.D{{Key: "$in", Value: []string{"project-a"}}}, bsonField(filter, "project_id"))

	unscoped := withProjectScope(context.Background(), "project_id", bson.D{{Key: "key", Value: "value"}})
	assert.Len(t, unscoped, 1)
}

func TestContainsPattern_QuotesUserInput(t *testing.T) {
	pattern := containsPattern("a.*b")
	assert.Equal(t, `.*a\.\*b.*`, pattern.Pattern)
	assert.Equal(t, "i", pattern.Options)
}

func TestRoleGetAll_KeepsTenantFilterWhenSearching(t *testing.T) {
	mockDB := test.SetupMockDB()
	ctx := scopedQueryContext()
	model := models.GetRoleModel()

	var captured bson.D
	mockDB.On("CountDocuments", ctx, model, mock.MatchedBy(func(filter interface{}) bool {
		document, ok := filter.(bson.D)
		if !ok {
			return false
		}
		captured = document
		return len(document) == 3
	}), mock.Anything).Return(int64(0), nil)
	cursor, err := mongo.NewCursorFromDocuments([]interface{}{}, nil, nil)
	require.NoError(t, err)
	mockDB.On("Find", ctx, model, mock.Anything, mock.Anything).Return(cursor, nil)

	_, err = NewRoleRepository(mockDB).GetAll(ctx, sdk.RoleQuery{
		SearchQuery: "a.*b",
		ProjectIds:  []string{"project-b"},
	})
	require.NoError(t, err)

	assert.Equal(t, true, bsonField(captured, model.EnabledKey))
	assert.Equal(t, bson.D{{Key: "$in", Value: []string{"project-a"}}}, bsonField(captured, model.ProjectIdKey))
	assert.NotNil(t, bsonField(captured, "$or"))
}
