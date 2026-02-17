package routes

import (
	"errors"
	"net/http"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

func (h *IAMHandler) RegisterProjectRoutes(router fiber.Router) {
	v1 := router.Group("/project/v1")
	v1.Post("/", h.CreateProject)
	v1.Get("/", h.GetAllProjects)
	v1.Get("/:id", h.GetProjectById)
	v1.Put("/:id", h.UpdateProject)
}

func (h *IAMHandler) CreateProject(c *fiber.Ctx) error {
	payload := new(sdk.Project)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.ProjectResponse{Success: false, Message: err.Error()})
	}
	if err := h.prv.S.AdminUC.CreateProject(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ProjectResponse{Success: false, Message: err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(sdk.ProjectResponse{Success: true, Data: payload})
}

func (h *IAMHandler) GetAllProjects(c *fiber.Ctx) error {
	projects, err := h.prv.S.AdminUC.GetAllProjects(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ProjectsResponse{Success: false, Message: err.Error()})
	}
	// Note: GetAllProjects returns []entity.Project, and sdk.ProjectsResponse expects []sdk.Project.
	// Since sdk.Project is an alias to entity.Project, we can pass it directly or cast if needed.
	return c.JSON(sdk.ProjectsResponse{Success: true, Data: projects})
}

func (h *IAMHandler) GetProjectById(c *fiber.Ctx) error {
	id := c.Params("id")
	project, err := h.prv.S.AdminUC.GetProject(c.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdk.ErrProjectNotFound) {
			status = http.StatusNotFound
		}
		return c.Status(status).JSON(sdk.ProjectResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ProjectResponse{Success: true, Data: project})
}

func (h *IAMHandler) UpdateProject(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.Project)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.ProjectResponse{Success: false, Message: err.Error()})
	}
	payload.Id = id
	if err := h.prv.S.AdminUC.UpdateProject(c.Context(), payload); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.ProjectResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ProjectResponse{Success: true, Data: payload})
}
