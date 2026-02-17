package sdk

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
)

var (
	// ErrResourceNotFound is returned when a requested resource cannot be found.
	ErrResourceNotFound = entity.ErrResourceNotFound
)

// Resource represents a resource in the Go IAM system.
type Resource = entity.Resource

// ResourceQuery represents search and filtering criteria for resource queries.
// This is used for listing resources with various filters and pagination.
type ResourceQuery struct {
	ProjectIds  []string `json:"project_ids,omitempty"` // Filter by specific project IDs
	Name        string   `json:"name,omitempty"`        // Filter by resource name (partial match)
	Description string   `json:"description,omitempty"` // Filter by resource description (partial match)
	Key         string   `json:"key,omitempty"`         // Filter by resource key (partial match)
	Skip        int64    `json:"skip"`                  // Number of records to skip (pagination)
	Limit       int64    `json:"limit"`                 // Maximum number of records to return
}

// ResourceResponse represents an API response containing a single resource.
type ResourceResponse struct {
	Success bool      `json:"success"`        // Indicates if the operation was successful
	Message string    `json:"message"`        // Human-readable message about the operation
	Data    *Resource `json:"data,omitempty"` // The resource data (present only on success)
}

// ResourceList represents a paginated list of resources with metadata.
type ResourceList struct {
	Resources []Resource `json:"resources"` // Array of resource objects
	Total     int64      `json:"total"`     // Total number of resources matching the query (before pagination)
	Skip      int64      `json:"skip"`      // Number of records skipped
	Limit     int64      `json:"limit"`     // Maximum number of records returned
}

// ResourcesResponse represents an API response containing a list of resources.
type ResourcesResponse struct {
	Success bool          `json:"success"`        // Indicates if the operation was successful
	Message string        `json:"message"`        // Human-readable message about the operation
	Data    *ResourceList `json:"data,omitempty"` // The paginated resource list data
}
