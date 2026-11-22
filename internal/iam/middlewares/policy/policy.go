// Package policy provides request-level authorization checks for the IAM API.
package policy

import (
	"encoding/json"
	"net/http"

	iamMiddlewares "github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

type projectScopedBody struct {
	ProjectID string `json:"project_id"`
	ProjectId string `json:"projectId"`
}

// Enforce rejects requests that lack a verified project scope or try to read
// or write a project outside that scope. By-ID authorization happens in the
// service layer once the stored resource is available.
func Enforce(c *fiber.Ctx) error {
	user := iamMiddlewares.GetUser(c.Context())
	if user == nil {
		// Internal calls and the explicit local-dev bypass do not carry an
		// authenticated identity.
		return c.Next()
	}
	if err := iamMiddlewares.RequireProjectScope(c.Context()); err != nil {
		return forbidden(c, err.Error())
	}
	for _, projectID := range requestedProjects(c) {
		if err := iamMiddlewares.AuthorizeProject(c.Context(), projectID); err != nil {
			return forbidden(c, err.Error())
		}
	}
	return c.Next()
}

func requestedProjects(c *fiber.Ctx) []string {
	ids := []string{}
	for _, key := range []string{"project_id", "projectId"} {
		if value := c.Query(key); value != "" {
			ids = append(ids, value)
		}
	}
	if len(c.Body()) == 0 {
		return ids
	}
	var body projectScopedBody
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return ids
	}
	if body.ProjectID != "" {
		ids = append(ids, body.ProjectID)
	}
	if body.ProjectId != "" {
		ids = append(ids, body.ProjectId)
	}
	return ids
}

func forbidden(c *fiber.Ctx, message string) error {
	return c.Status(http.StatusForbidden).JSON(sdk.UserResponse{
		Success: false,
		Message: message,
	})
}
