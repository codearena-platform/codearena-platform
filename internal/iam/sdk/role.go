package sdk

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
)

var ErrRoleNotFound = entity.ErrRoleNotFound

type Role = entity.Role
type Resources = entity.Resources

// RoleQuery represents search and filtering criteria for role queries.
// This is used for listing roles with various filters and pagination.
type RoleQuery struct {
	ProjectIds  []string `json:"project_ids"`  // Filter by specific project IDs
	SearchQuery string   `json:"search_query"` // Text search across role fields
	Skip        int64    `json:"skip"`         // Number of records to skip (pagination)
	Limit       int64    `json:"limit"`        // Maximum number of records to return
}

// RoleResponse represents an API response containing a single role.
type RoleResponse struct {
	Success bool   `json:"success"`        // Indicates if the operation was successful
	Message string `json:"message"`        // Human-readable message about the operation
	Data    *Role  `json:"data,omitempty"` // The role data (present only on success)
}

// RoleList represents a paginated list of roles with metadata.
type RoleList struct {
	Roles []Role `json:"roles"` // Array of role objects
	Total int64  `json:"total"` // Total number of roles matching the query (before pagination)
	Skip  int64  `json:"skip"`  // Number of records skipped
	Limit int64  `json:"limit"` // Maximum number of records returned
}

// RoleListResponse represents an API response containing a list of roles.
type RoleListResponse struct {
	Success bool      `json:"success"`        // Indicates if the operation was successful
	Message string    `json:"message"`        // Human-readable message about the operation
	Data    *RoleList `json:"data,omitempty"` // The paginated role list data
}
