package routes

import (
	"net/http"

	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

func (h *IAMHandler) RegisterMeRoutes(router fiber.Router) {
	router.Get("/me", h.GetMe)
	router.Get("/me/dashboard", h.AuthClientCheck, h.prv.AM.DashboardUser, h.GetDashboardMe)
}

func (h *IAMHandler) GetMe(c *fiber.Ctx) error {
	user := middlewares.GetUser(c.Context())
	return c.Status(http.StatusOK).JSON(sdk.UserResponse{
		Success: true,
		Message: "User fetched successfully",
		Data:    user,
	})
}

func (h *IAMHandler) AuthClientCheck(c *fiber.Ctx) error {
	if h.prv.AuthClient == nil {
		res := sdk.DashboardUserResponse{Success: true, Message: "auth is not setup yet."}
		res.Data.Setup.ClientAdded = false
		return c.Status(http.StatusOK).JSON(res)
	}
	return c.Next()
}

func (h *IAMHandler) GetDashboardMe(c *fiber.Ctx) error {
	user := middlewares.GetUser(c.Context())
	res := sdk.DashboardUserResponse{Success: true, Message: "User fetched successfully"}
	res.Data.Setup.ClientAdded = true
	res.Data.Setup.ClientId = h.prv.AuthClient.Id
	res.Data.User = user
	return c.Status(http.StatusOK).JSON(res)
}
