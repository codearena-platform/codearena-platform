// Package projects provides project scoping metadata for the Go IAM system.
// Project scopes are derived from the authenticated identity; request headers
// are never accepted as a source of authority.
package projects

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/project"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Middlewares provides project-scoping middleware functionality.
type Middlewares struct {
	projectSvc project.Service
}

// NewMiddlewares creates a new project middleware instance.
func NewMiddlewares(projectSvc project.Service) *Middlewares {
	return &Middlewares{projectSvc: projectSvc}
}

// Projects seeds the project scope from an already-authenticated user.
// The legacy X-Project-Ids header is ignored: the authentication middleware
// derives the authoritative scope from the verified identity and only uses the
// header to narrow it.
func (m Middlewares) Projects(c *fiber.Ctx) error {
	if c.Get("X-Project-Ids") != "" {
		log.Warn("X-Project-Ids is deprecated and no longer grants project access")
	}
	if user := middlewares.GetUser(c.Context()); user != nil && user.ProjectId != "" {
		c.Context().SetUserValue(sdk.ProjectsTypeVal, []string{user.ProjectId})
	}
	return c.Next()
}
