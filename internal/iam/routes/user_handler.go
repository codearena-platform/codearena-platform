package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/docs"
	"github.com/gofiber/fiber/v2"
)

var userTags = []string{"User"}

func (h *IAMHandler) RegisterUserRoutes(router fiber.Router) {
	v1 := router.Group("/user/v1")

	// Register Handlers
	v1.Post("/", h.CreateUser)
	v1.Get("/:id", h.GetUserById)
	v1.Get("/", h.GetAllUsers)
	v1.Put("/:id", h.UpdateUser)
	v1.Put("/:id/roles", h.UpdateUserRoles)
	v1.Put("/:id/policies", h.UpdateUserPolicies)
	v1.Put("/:id/transfer-ownership/:oldId", h.TransferOwnership)
	v1.Put("/:sourceId/copy-resources/:targetId", h.CopyUserResources)

	// Register Documentation
	h.registerUserDocs("/api/v1/iam/user/v1")
}

func (h *IAMHandler) CreateUser(c *fiber.Ctx) error {
	payload := new(sdk.User)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request. %v", err),
		})
	}
	err := h.prv.S.IdentityUC.CreateUser(c.Context(), payload)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create user. %v", err),
		})
	}
	return c.Status(http.StatusCreated).JSON(sdk.UserResponse{
		Success: true,
		Message: "User created successfully",
		Data:    payload,
	})
}

func (h *IAMHandler) GetUserById(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.prv.S.IdentityUC.GetUserById(c.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), sdk.ErrUserNotFound.Error()) {
			return c.Status(http.StatusNotFound).JSON(sdk.UserResponse{
				Success: false,
				Message: "User not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get user. %v", err),
		})
	}
	return c.JSON(sdk.UserResponse{
		Success: true,
		Message: "User fetched successfully",
		Data:    user,
	})
}

func (h *IAMHandler) GetAllUsers(c *fiber.Ctx) error {
	query := sdk.UserQuery{
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
	users, err := h.prv.S.IdentityUC.GetAllUsers(c.Context(), query)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get users. %v", err),
		})
	}
	return c.JSON(sdk.UserListResponse{
		Success: true,
		Message: "Users fetched successfully",
		Data:    users,
	})
}

func (h *IAMHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.User)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request. %v", err),
		})
	}
	payload.Id = id
	err := h.prv.S.IdentityUC.UpdateUser(c.Context(), payload)
	if err != nil {
		if errors.Is(err, sdk.ErrUserNotFound) {
			return c.Status(http.StatusNotFound).JSON(sdk.UserResponse{
				Success: false,
				Message: "User not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("failed to update user. %v", err),
		})
	}
	return c.JSON(sdk.UserResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    payload,
	})
}

func (h *IAMHandler) UpdateUserRoles(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.UserRoleUpdate)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request. %v", err),
		})
	}

	for _, roleId := range payload.ToBeRemoved {
		if err := h.prv.S.IdentityUC.RemoveRoleFromUser(c.Context(), id, roleId); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{
				Success: false,
				Message: fmt.Sprintf("failed to remove role %s. %v", roleId, err),
			})
		}
	}

	for _, roleId := range payload.ToBeAdded {
		if err := h.prv.S.IdentityUC.AddRoleToUser(c.Context(), id, roleId); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{
				Success: false,
				Message: fmt.Sprintf("failed to add role %s. %v", roleId, err),
			})
		}
	}

	return c.JSON(sdk.UserResponse{Success: true, Message: "User roles updated successfully"})
}

func (h *IAMHandler) UpdateUserPolicies(c *fiber.Ctx) error {
	id := c.Params("id")
	payload := new(sdk.UserPolicyUpdate)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.UserResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request. %v", err),
		})
	}

	if err := h.prv.S.IdentityUC.RemovePolicyFromUser(c.Context(), id, payload.ToBeRemoved); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{Success: false, Message: err.Error()})
	}
	if err := h.prv.S.IdentityUC.AddPolicyToUser(c.Context(), id, payload.ToBeAdded); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{Success: false, Message: err.Error()})
	}

	return c.JSON(sdk.UserResponse{Success: true, Message: "User policies updated successfully"})
}

func (h *IAMHandler) TransferOwnership(c *fiber.Ctx) error {
	id := c.Params("id")
	oldId := c.Params("oldId")
	if err := h.prv.S.IdentityUC.TransferOwnership(c.Context(), id, oldId); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.UserResponse{Success: true, Message: "User ownership transferred successfully"})
}

func (h *IAMHandler) CopyUserResources(c *fiber.Ctx) error {
	sourceId := c.Params("sourceId")
	targetId := c.Params("targetId")
	if err := h.prv.S.IdentityUC.CopyUserResources(c.Context(), sourceId, targetId); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(sdk.UserResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.UserResponse{Success: true, Message: "User resources copied successfully"})
}

func (h *IAMHandler) registerUserDocs(basePath string) {
	docs.RegisterApi(docs.ApiWrapper{
		Path: basePath + "/", Method: http.MethodPost, Name: "Create User", Tags: userTags,
		RequestBody: &docs.ApiRequestBody{Content: new(sdk.User)},
		Response:    &docs.ApiResponse{Content: new(sdk.UserResponse)},
	})
	// ... we can add more docs as needed, keeping it flat makes it easier to manage
}
