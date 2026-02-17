package sdk

import (
	"net/http"

	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/gofiber/fiber/v2"
)

var ErrAuthProviderNotFound = entity.ErrAuthProviderNotFound

type AuthProviderType = entity.AuthProviderType

const (
	AuthProviderTypeGoogle    = entity.AuthProviderTypeGoogle
	AuthProviderTypeMicrosoft = entity.AuthProviderTypeMicrosoft
	AuthProviderTypeGitHub    = entity.AuthProviderTypeGitHub
	AuthProviderTypeOIDC      = entity.AuthProviderTypeOIDC
)

type AuthProvider = entity.AuthProvider
type AuthProviderParam = entity.AuthProviderParam

// AuthProviderResponse represents an API response containing a single authentication provider.
type AuthProviderResponse struct {
	Success bool          `json:"success"` // Indicates if the operation was successful
	Message string        `json:"message"` // Human-readable message about the operation
	Data    *AuthProvider `json:"data"`    // The auth provider data (present only on success)
}

// NewErrorAuthProviderResponse creates a new error response for auth provider operations.
func NewErrorAuthProviderResponse(msg string, status int, c *fiber.Ctx) error {
	return c.Status(status).JSON(ClientResponse{
		Success: false,
		Message: msg,
	})
}

// AuthProviderBadRequest returns a 400 Bad Request error response for auth provider operations.
func AuthProviderBadRequest(msg string, c *fiber.Ctx) error {
	return NewErrorAuthProviderResponse(msg, http.StatusBadRequest, c)
}

// AuthProviderNotFound returns a 404 Not Found error response for auth provider operations.
func AuthProviderNotFound(msg string, c *fiber.Ctx) error {
	return NewErrorAuthProviderResponse(msg, http.StatusNotFound, c)
}

// AuthProviderInternalServerError returns a 500 Internal Server Error response for auth provider operations.
func AuthProviderInternalServerError(msg string, c *fiber.Ctx) error {
	return NewErrorAuthProviderResponse(msg, http.StatusInternalServerError, c)
}

// AuthProvidersResponse represents an API response containing a list of authentication providers.
type AuthProvidersResponse struct {
	Success bool           `json:"success"` // Indicates if the operation was successful
	Message string         `json:"message"` // Human-readable message about the operation
	Data    []AuthProvider `json:"data"`    // Array of auth provider data
}

// NewErrorAuthProvidersResponse creates a new error response for auth providers list operations.
func NewErrorAuthProvidersResponse(msg string, status int, c *fiber.Ctx) error {
	return c.Status(status).JSON(AuthProvidersResponse{
		Success: false,
		Message: msg,
	})
}

// AuthProvidersInternalServerError returns a 500 Internal Server Error response for auth providers list operations.
func AuthProvidersInternalServerError(msg string, c *fiber.Ctx) error {
	return NewErrorAuthProvidersResponse(msg, http.StatusInternalServerError, c)
}

// ServiceProvider interface defines the contract for external authentication service implementations.
// This interface must be implemented by each supported authentication provider (Google, Microsoft, GitHub, etc.)
// to provide OAuth2/OIDC functionality.
type ServiceProvider = entity.ServiceProvider

// AuthProviderQueryParams represents query parameters for filtering authentication providers.
type AuthProviderQueryParams struct {
	ProjectIds []string `json:"project_id"` // Filter providers by project IDs
}
