package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

var roleTags = []string{"Role"}

func (h *IAMHandler) RegisterRoleRoutes(router fiber.Router) {
	v1 := router.Group("/role/v1")

	v1.Post("/", h.CreateRole)
	v1.Get("/", h.SearchRoles)
	v1.Get("/:id", h.GetRoleById)
	v1.Put("/:id", h.UpdateRole)
}

func (h *IAMHandler) CreateRole(c *fiber.Ctx) error {
	payload := new(sdk.Role)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.RoleResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request. %v", err),
		})
	}
	err := h.prv.S.Role.Create(c.Context(), payload)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.RoleResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create role. %v", err),
		})
	}
	return c.Status(http.StatusCreated).JSON(sdk.RoleResponse{
		Success: true,
		Message: "Role created successfully",
		Data:    payload,
	})
}

func (h *IAMHandler) SearchRoles(c *fiber.Ctx) error {
	query := sdk.RoleQuery{
		SearchQuery: c.Query("query"),
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
	roles, err := h.prv.S.Role.GetAll(c.Context(), query)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.RoleListResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(sdk.RoleListResponse{
		Success: true,
		Data:    roles,
	})
}

func (h *IAMHandler) GetRoleById(c *fiber.Ctx) error {
	id := c.Params("id")
	role, err := h.prv.S.Role.GetById(c.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdk.ErrRoleNotFound) {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(sdk.RoleResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(sdk.RoleResponse{
		Success: true,
		Data:    role,
	})
}

func (h *IAMHandler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.Role)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.RoleResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	payload.Id = id
	err := h.prv.S.Role.Update(c.Context(), payload)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.RoleResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(sdk.RoleResponse{
		Success: true,
		Data:    payload,
	})
}
