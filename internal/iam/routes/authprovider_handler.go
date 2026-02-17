package routes

import (
	"errors"
	"net/http"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

func (h *IAMHandler) RegisterAuthProviderRoutes(router fiber.Router) {
	v1 := router.Group("/authprovider/v1")
	v1.Post("/", h.CreateAuthProvider)
	v1.Get("/", h.GetAllAuthProviders)
	v1.Get("/:id", h.GetAuthProviderById)
	v1.Put("/:id", h.UpdateAuthProvider)
}

func (h *IAMHandler) CreateAuthProvider(c *fiber.Ctx) error {
	payload := new(sdk.AuthProvider)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.AuthProviderResponse{Success: false, Message: err.Error()})
	}
	if err := h.prv.S.AuthUC.CreateAuthProvider(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.AuthProviderResponse{Success: false, Message: err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(sdk.AuthProviderResponse{Success: true, Data: payload})
}

func (h *IAMHandler) GetAllAuthProviders(c *fiber.Ctx) error {
	providers, err := h.prv.S.AuthUC.GetAllAuthProviders(c.Context(), sdk.AuthProviderQueryParams{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.AuthProvidersResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.AuthProvidersResponse{Success: true, Data: providers})
}

func (h *IAMHandler) GetAuthProviderById(c *fiber.Ctx) error {
	id := c.Params("id")
	provider, err := h.prv.S.AuthUC.GetAuthProvider(c.Context(), id, false)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdk.ErrAuthProviderNotFound) {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(sdk.AuthProviderResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.AuthProviderResponse{Success: true, Data: provider})
}

func (h *IAMHandler) UpdateAuthProvider(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.AuthProvider)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.AuthProviderResponse{Success: false, Message: err.Error()})
	}
	payload.Id = id
	if err := h.prv.S.AuthUC.UpdateAuthProvider(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.AuthProviderResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.AuthProviderResponse{Success: true, Data: payload})
}
