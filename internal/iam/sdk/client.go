package sdk

import (
	"net/http"

	"github.com/codearena-platform/codearena-platform/internal/iam/entity"

	"github.com/gofiber/fiber/v2"
)

// ErrClientNotFound is returned when a requested OAuth2 client cannot be found.
var ErrClientNotFound = entity.ErrClientNotFound

// Client represents an OAuth2 client in the Go IAM system.
type Client = entity.Client

// ClientResponse represents an API response containing a single OAuth2 client.
type ClientResponse struct {
	Success bool    `json:"success"`        // Indicates if the operation was successful
	Message string  `json:"message"`        // Human-readable message about the operation
	Data    *Client `json:"data,omitempty"` // The client data (present only on success)
}

// NewErrorClientResponse creates a new error response for client operations.
func NewErrorClientResponse(msg string, status int, c *fiber.Ctx) error {
	return c.Status(status).JSON(ClientResponse{
		Success: false,
		Message: msg,
	})
}

// ClientBadRequest returns a 400 Bad Request error response for client operations.
func ClientBadRequest(msg string, c *fiber.Ctx) error {
	return NewErrorClientResponse(msg, http.StatusBadRequest, c)
}

// ClientNotFound returns a 404 Not Found error response for client operations.
func ClientNotFound(msg string, c *fiber.Ctx) error {
	return NewErrorClientResponse(msg, http.StatusNotFound, c)
}

// ClientInternalServerError returns a 500 Internal Server Error response for client operations.
func ClientInternalServerError(msg string, c *fiber.Ctx) error {
	return NewErrorClientResponse(msg, http.StatusInternalServerError, c)
}

// ClientsResponse represents an API response containing a list of OAuth2 clients.
type ClientsResponse struct {
	Success bool     `json:"success"` // Indicates if the operation was successful
	Message string   `json:"message"` // Human-readable message about the operation
	Data    []Client `json:"data"`    // Array of client data
}

// NewErrorClientsResponse creates a new error response for clients list operations.
func NewErrorClientsResponse(msg string, status int, c *fiber.Ctx) error {
	return c.Status(status).JSON(ClientsResponse{
		Success: false,
		Message: msg,
	})
}

// ClientsInternalServerError returns a 500 Internal Server Error response for clients list operations.
func ClientsInternalServerError(msg string, c *fiber.Ctx) error {
	return NewErrorClientsResponse(msg, http.StatusInternalServerError, c)
}

// ClientQueryParams represents query parameters for filtering and sorting OAuth2 clients.
type ClientQueryParams struct {
	ProjectIds      []string `json:"project_id"`         // Filter clients by project IDs
	GoIamClient     bool     `json:"go_iam_client"`      // Filter by Go-IAM internal clients
	SortByUpdatedAt bool     `json:"sort_by_updated_at"` // Sort results by update timestamp
}
