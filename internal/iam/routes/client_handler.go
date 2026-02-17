package routes

import (
	"errors"
	"net/http"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

func (h *IAMHandler) RegisterClientRoutes(router fiber.Router) {
	v1 := router.Group("/client/v1")
	v1.Post("/", h.CreateClient)
	v1.Get("/", h.GetAllClients)
	v1.Get("/:id", h.GetClientById)
	v1.Put("/:id", h.UpdateClient)
	v1.Put("/:id/regenerate-secret", h.RegenerateClientSecret)
}

func (h *IAMHandler) CreateClient(c *fiber.Ctx) error {
	payload := new(sdk.Client)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.ClientResponse{Success: false, Message: err.Error()})
	}
	if err := h.prv.S.Clients.Create(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ClientResponse{Success: false, Message: err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(sdk.ClientResponse{Success: true, Data: payload})
}

func (h *IAMHandler) GetAllClients(c *fiber.Ctx) error {
	clients, err := h.prv.S.Clients.GetAll(c.Context(), sdk.ClientQueryParams{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ClientsResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ClientsResponse{Success: true, Data: clients})
}

func (h *IAMHandler) GetClientById(c *fiber.Ctx) error {
	id := c.Params("id")
	client, err := h.prv.S.Clients.Get(c.Context(), id, false)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdk.ErrClientNotFound) {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(sdk.ClientResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ClientResponse{Success: true, Data: client})
}

func (h *IAMHandler) UpdateClient(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.Client)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.ClientResponse{Success: false, Message: err.Error()})
	}
	payload.Id = id
	if err := h.prv.S.Clients.Update(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ClientResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ClientResponse{Success: true, Data: payload})
}

func (h *IAMHandler) RegenerateClientSecret(c *fiber.Ctx) error {
	id := c.Params("id")
	client, err := h.prv.S.Clients.RegenerateSecret(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ClientResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ClientResponse{Success: true, Message: "Client secret regenerated", Data: client})
}
