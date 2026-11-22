// Package auth provides authentication middleware for the IAM system.
// It handles user authentication via Bearer tokens and integrates with
// the IAM authentication service to validate and extract user information.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/auth"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/client"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/httpclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Middlewares provides authentication middleware functionality.
// It encapsulates the authentication service, client service, and IAM client
// needed for user authentication and authorization.
type Middlewares struct {
	authSvc    auth.Service   // Authentication service for token validation
	clientSvc  client.Service // Client service for client operations
	AuthClient *sdk.Client    // IAM client configuration
}

// NewMiddlewares creates a new authentication middleware instance.
// It initializes the middleware with the required services and retrieves
// the IAM client configuration for authentication operations.
//
// Parameters:
//   - authSvc: Authentication service for token validation
//   - clientSvc: Client service for client operations
//
// Returns:
//   - *Middlewares: Configured middleware instance
//   - error: Error if IAM client retrieval fails
func NewMiddlewares(authSvc auth.Service, clientSvc client.Service) (*Middlewares, error) {
	authClient, err := httpclient.GetIAMClient(clientSvc)
	if err != nil {
		return nil, err
	}
	return &Middlewares{
		authSvc:    authSvc,
		clientSvc:  clientSvc,
		AuthClient: authClient,
	}, nil
}

// User is a Fiber middleware that authenticates users via Bearer token.
// It extracts the Authorization header, validates the token, and stores
// the authenticated user in the request context. If authentication fails,
// it returns a 401 Unauthorized response.
//
// Usage:
//
//	app.Use(authMiddleware.User)
//
// Parameters:
//   - c: Fiber context containing the HTTP request
//
// Returns:
//   - error: nil on success, HTTP error response on authentication failure
func (m *Middlewares) User(c *fiber.Ctx) error {
	if m.AuthClient == nil {
		// Without an IAM client there is no way to validate a token. Fail closed
		// unless the deployment explicitly opts into insecure development mode.
		if os.Getenv("IAM_ALLOW_INSECURE") != "true" {
			return c.Status(http.StatusUnauthorized).JSON(sdk.UserResponse{
				Success: false,
				Message: "authentication is not configured",
			})
		}
		setInsecureProjects(c)
		return c.Next()
	}
	// This middleware can be used to check if the user is authenticated
	user, err := m.GetUser(c)
	if err != nil {
		log.Warnw("failed to fetch user", "error", err)
		return c.Status(http.StatusUnauthorized).JSON(sdk.UserResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	c.Context().SetUserValue(sdk.UserTypeVal, user)
	return c.Next()
}

// DashboardUser is a specialized Fiber middleware for dashboard authentication.
// It performs user authentication similar to User middleware but returns
// dashboard-specific response format including setup information about
// the IAM client configuration.
//
// Usage:
//
//	app.Use(authMiddleware.DashboardUser)
//
// Parameters:
//   - c: Fiber context containing the HTTP request
//
// Returns:
//   - error: nil on success, HTTP error response with dashboard format on failure
func (m *Middlewares) DashboardUser(c *fiber.Ctx) error {
	res := &sdk.DashboardUserResponse{
		Success: false,
	}
	if m.AuthClient == nil {
		if os.Getenv("IAM_ALLOW_INSECURE") != "true" {
			res.Message = "authentication is not configured"
			return c.Status(http.StatusUnauthorized).JSON(res)
		}
		res.Data.Setup.ClientAdded = false
		setInsecureProjects(c)
		return c.Next()
	}
	res.Data.Setup.ClientAdded = true
	res.Data.Setup.ClientId = m.AuthClient.Id
	// This middleware can be used to check if the user is authenticated
	user, err := m.GetUser(c)
	if err != nil {
		log.Warnw("failed to fetch user", "error", err)
		res.Message = err.Error()
		return c.Status(http.StatusUnauthorized).JSON(res)
	}

	c.Context().SetUserValue(sdk.UserTypeVal, user)
	return c.Next()
}

// setInsecureProjects is only used when IAM_ALLOW_INSECURE=true. The header is
// never trusted when a verified token is available.
func setInsecureProjects(c *fiber.Ctx) {
	log.Warn("IAM_ALLOW_INSECURE=true: project scope taken from X-Project-Ids (development only)")
	csv := c.Get("X-Project-Ids")
	ids := []string{}
	if csv != "" {
		for _, id := range strings.Split(csv, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				ids = append(ids, trimmed)
			}
		}
	}
	c.Context().SetUserValue(sdk.ProjectsTypeVal, ids)
}

// GetUser extracts and validates the user from the Authorization header.
// This is a helper method used by the middleware functions to perform
// the actual token validation and user retrieval.
//
// Parameters:
//   - c: Fiber context containing the HTTP request with Authorization header
//
// Returns:
//   - *sdk.User: Authenticated user information
//   - error: Error if token is missing, invalid, or authentication fails
func (m *Middlewares) GetUser(c *fiber.Ctx) (*sdk.User, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, errors.New("`Authorization` not found in header")
	}

	// Extract the token from the header
	token := authHeader[len("Bearer "):]
	if token == "" {
		return nil, errors.New("`Bearer` token not found in header")
	}

	user, err := m.authSvc.GetIdentity(c.Context(), token, c.QueryBool("force_fetch", false))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch identity: %w", err)
	}
	// The verified identity is the only source of project authority. The header
	// may narrow the scope to a subset, but never add a project the user does
	// not belong to.
	allowed := []string{}
	if user.ProjectId != "" {
		allowed = append(allowed, user.ProjectId)
	}
	if csv := c.Get("X-Project-Ids"); csv != "" {
		narrowed := []string{}
		for _, requested := range strings.Split(csv, ",") {
			for _, projectID := range allowed {
				if strings.TrimSpace(requested) == projectID {
					narrowed = append(narrowed, projectID)
					break
				}
			}
		}
		if len(narrowed) == 0 {
			return nil, errors.New("no authorized project scope in X-Project-Ids")
		}
		allowed = narrowed
	}
	c.Context().SetUserValue(sdk.ProjectsTypeVal, allowed)
	return user, nil
}
