package repository

import (
	"context"
	"regexp"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// projectIDsForQuery returns the verified project scope for authenticated
// callers. Callers that have no authenticated identity (internal maintenance
// and event handlers) may supply their own IDs.
func projectIDsForQuery(ctx context.Context, requested []string) []string {
	if middlewares.GetUser(ctx) == nil {
		return requested
	}
	return middlewares.GetProjects(ctx)
}

// withProjectScope restricts a maintenance filter to the caller's verified
// project scope when one is present.
func withProjectScope(ctx context.Context, projectIDKey string, filter bson.D) bson.D {
	if middlewares.GetUser(ctx) == nil {
		return filter
	}
	return append(filter, bson.E{Key: projectIDKey, Value: bson.D{{Key: "$in", Value: middlewares.GetProjects(ctx)}}})
}

// containsPattern builds a case-insensitive substring regex. User input is
// quoted so it cannot change the meaning of the query.
func containsPattern(value string) primitive.Regex {
	return primitive.Regex{Pattern: ".*" + regexp.QuoteMeta(value) + ".*", Options: "i"}
}
