// Package middlewares provides utility functions for extracting and managing
// authentication and authorization metadata from request contexts.
// It handles user information, project access, and metadata operations
// used throughout the Go IAM system.
package middlewares

import (
	"context"
	"errors"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
)

var (
	// ErrProjectScopeMissing is returned when an authenticated request has no
	// verified project scope.
	ErrProjectScopeMissing = errors.New("no authorized project scope")
	// ErrCrossTenantAccess is returned when a resource belongs to a project the
	// caller is not scoped to.
	ErrCrossTenantAccess = errors.New("resource belongs to another project")
)

// GetProjects extracts the verified project scope from the given context.
// Values are derived from the authenticated identity by the auth middleware.
//
// Parameters:
//   - ctx: The context containing project information
//
// Returns:
//   - []string: Slice of project IDs if found, nil if not present or invalid type
func GetProjects(ctx context.Context) []string {
	val, ok := ctx.Value(sdk.ProjectsTypeVal).([]string)
	if !ok {
		return nil
	}
	return val
}

// GetUser extracts the authenticated user from the given context.
// This function retrieves user information that was previously stored in the context
// by the authentication middleware after successful token validation.
//
// Parameters:
//   - ctx: The context containing user information
//
// Returns:
//   - *sdk.User: Pointer to the authenticated user if found, nil if not present or invalid type
func GetUser(ctx context.Context) *sdk.User {
	user := ctx.Value(sdk.UserTypeVal)
	if user == nil {
		return nil
	}
	authUser, ok := user.(*sdk.User)
	if !ok {
		return nil
	}
	return authUser
}

// GetMetadata creates a consolidated metadata object containing both user and project information
// from the context. This function combines the results of GetUser and GetProjects for
// convenient access to all authentication and authorization metadata.
//
// Parameters:
//   - ctx: The context containing user and project information
//
// Returns:
//   - sdk.Metadata: Metadata object containing user and project IDs
func GetMetadata(ctx context.Context) sdk.Metadata {
	return sdk.Metadata{
		User:       GetUser(ctx),
		ProjectIds: GetProjects(ctx),
	}
}

// RequireProjectScope rejects authenticated requests without a verified
// project scope. Requests without an authenticated user are internal service
// calls and are allowed to operate across projects.
func RequireProjectScope(ctx context.Context) error {
	if GetUser(ctx) == nil {
		return nil
	}
	if len(GetProjects(ctx)) == 0 {
		return ErrProjectScopeMissing
	}
	return nil
}

// AuthorizeProject checks that an authenticated caller is scoped to the given
// project. It must be called on by-ID reads and on writes that carry a project.
func AuthorizeProject(ctx context.Context, projectID string) error {
	if err := RequireProjectScope(ctx); err != nil {
		return err
	}
	if GetUser(ctx) == nil {
		// Internal service and event calls are allowed to cross projects.
		return nil
	}
	if projectID == "" {
		// Legacy unscoped documents are not specific to another tenant.
		return nil
	}
	for _, allowed := range GetProjects(ctx) {
		if allowed == projectID {
			return nil
		}
	}
	return ErrCrossTenantAccess
}

// AddMetadata stores user and project information into the context.
// This function is typically used by services or handlers to inject metadata
// into the context for downstream processing.
//
// Parameters:
//   - ctx: The context to store metadata in
//   - metadata: The metadata containing user and project information to store
//
// Returns:
//   - context.Context: New context with the metadata stored
func AddMetadata(ctx context.Context, metadata sdk.Metadata) context.Context {
	return context.WithValue(context.WithValue(ctx, sdk.ProjectsTypeVal, metadata.ProjectIds), sdk.UserTypeVal, metadata.User)
}
