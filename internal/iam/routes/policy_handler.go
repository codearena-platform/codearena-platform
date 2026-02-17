package routes

import (
	"net/http"
	"strconv"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

func (h *IAMHandler) RegisterPolicyRoutes(router fiber.Router) {
	v1 := router.Group("/policy/v1")
	v1.Get("/", h.GetAllPolicies)
}

func (h *IAMHandler) GetAllPolicies(c *fiber.Ctx) error {
	query := sdk.PolicyQuery{
		Query: c.Query("query"),
		Skip:  0,
		Limit: 10,
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
	policies, err := h.prv.S.AccessUC.GetAllPolicies(c.Context(), query)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.PoliciesResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.PoliciesResponse{Success: true, Data: *policies})
}
