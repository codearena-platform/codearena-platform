package sdk

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
)

var ErrPolicyNotFound = entity.ErrPolicyNotFound

type Policy = entity.Policy
type PolicyDefinition = entity.PolicyDefinition
type PolicyArgument = entity.PolicyArgument

// PolicyResponse represents an API response containing a single policy.
type PolicyResponse struct {
	Success bool    `json:"success"`        // Indicates if the operation was successful
	Message string  `json:"message"`        // Human-readable message about the operation
	Data    *Policy `json:"data,omitempty"` // The policy data (present only on success)
}

// PoliciesResponse represents an API response containing a list of policies.
type PoliciesResponse struct {
	Success bool       `json:"success"`        // Indicates if the operation was successful
	Message string     `json:"message"`        // Human-readable message about the operation
	Data    PolicyList `json:"data,omitempty"` // The paginated policy list data
}

// PolicyList represents a paginated list of policies with metadata.
type PolicyList struct {
	Policies []Policy `json:"policies"` // Array of policy objects
	Total    int      `json:"total"`    // Total number of policies matching the query (before pagination)
	Skip     int64    `json:"skip"`     // Number of records skipped
	Limit    int64    `json:"limit"`    // Maximum number of records returned
}

// PolicyQuery represents search and filtering criteria for policy queries.
// This is used for listing policies with text search and pagination.
type PolicyQuery struct {
	Query string `json:"query,omitempty"` // Text search query across policy fields
	Skip  int64  `json:"skip,omitempty"`  // Number of records to skip (pagination)
	Limit int64  `json:"limit,omitempty"` // Maximum number of records to return
}
