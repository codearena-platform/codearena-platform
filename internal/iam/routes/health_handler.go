package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthResponse struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message"`
	Data      HealthData `json:"data"`
	Timestamp string     `json:"timestamp"`
}

type HealthData struct {
	Status     string            `json:"status"`
	Version    string            `json:"version"`
	Uptime     string            `json:"uptime"`
	Components map[string]string `json:"components"`
}

func (h *IAMHandler) RegisterHealthRoutes(router fiber.Router) {
	router.Get("/health", h.HealthCheck)
}

func (h *IAMHandler) HealthCheck(c *fiber.Ctx) error {
	uptime := time.Since(h.startTime)
	components := make(map[string]string)

	if h.prv.D != nil {
		_, err := h.prv.S.Projects.GetAll(c.Context())
		if err != nil {
			components["database"] = "unhealthy"
		} else {
			components["database"] = "healthy"
		}
	}

	status := "healthy"
	for _, s := range components {
		if s == "unhealthy" {
			status = "unhealthy"
			break
		}
	}

	return c.JSON(HealthResponse{
		Success: true,
		Message: "Health check completed successfully",
		Data: HealthData{
			Status:     status,
			Version:    "1.0.0",
			Uptime:     uptime.String(),
			Components: components,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
