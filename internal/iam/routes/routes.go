package routes

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/providers"
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, prv *providers.Provider) {
	h := NewIAMHandler(prv)

	// Admin API (Requires Auth)
	api := app.Group("/api/v1/iam", prv.AM.User)
	h.RegisterProjectRoutes(api)
	h.RegisterClientRoutes(api)
	h.RegisterAuthProviderRoutes(api)
	h.RegisterUserRoutes(api)
	h.RegisterResourceRoutes(api)
	h.RegisterRoleRoutes(api)
	h.RegisterPolicyRoutes(api)
	h.RegisterMeRoutes(api)

	// Public API
	public := app.Group("/api/v1/iam")
	h.RegisterAuthRoutes(public)
	h.RegisterHealthRoutes(public)

	// Legacy /me redirect or direct access if needed
	h.RegisterMeRoutes(app)

	app.Static("/docs", "./docs")
}
