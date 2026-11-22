package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

func (h *IAMHandler) RegisterResourceRoutes(router fiber.Router) {
	v1 := router.Group("/resource/v1")
	v1.Post("/", h.CreateResource)
	// /search must be registered before /:id or Fiber treats "search" as an id.
	v1.Get("/search", h.SearchResources)
	v1.Get("/:id", h.GetResourceById)
	v1.Put("/:id", h.UpdateResource)
	v1.Delete("/:id", h.DeleteResource)
}

func (h *IAMHandler) CreateResource(c *fiber.Ctx) error {
	payload := new(sdk.Resource)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.ResourceResponse{Success: false, Message: err.Error()})
	}
	if err := h.prv.S.AccessUC.CreateResource(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ResourceResponse{Success: false, Message: err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(sdk.ResourceResponse{Success: true, Data: payload})
}

func (h *IAMHandler) GetResourceById(c *fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.prv.S.AccessUC.GetResource(c.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdk.ErrResourceNotFound) {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(sdk.ResourceResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ResourceResponse{Success: true, Data: res})
}

func (h *IAMHandler) SearchResources(c *fiber.Ctx) error {
	query := sdk.ResourceQuery{
		Name:        c.Query("name"),
		Description: c.Query("description"),
		Key:         c.Query("key"),
		Skip:        0,
		Limit:       10,
	}
	if skip := c.Query("skip"); skip != "" {
		if val, err := strconv.ParseInt(skip, 10, 64); err == nil {
			query.Skip = val
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
			query.Limit = val
		}
	}
	res, err := h.prv.S.AccessUC.GetAllResources(c.Context(), query)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ResourcesResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ResourcesResponse{Success: true, Data: res})
}

func (h *IAMHandler) UpdateResource(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.Resource)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.ResourceResponse{Success: false, Message: err.Error()})
	}
	payload.ID = id
	if err := h.prv.S.AccessUC.UpdateResource(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ResourceResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ResourceResponse{Success: true, Data: payload})
}

func (h *IAMHandler) DeleteResource(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.prv.S.AccessUC.DeleteResource(c.Context(), id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ResourceResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ResourceResponse{Success: true, Message: "Resource deleted successfully"})
}
