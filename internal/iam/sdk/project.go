package sdk

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
)

var ErrProjectNotFound = entity.ErrProjectNotFound

type Project = entity.Project

// ProjectResponse represents an API response containing a single project.
type ProjectResponse struct {
	Success bool     `json:"success"`        // Indicates if the operation was successful
	Message string   `json:"message"`        // Human-readable message about the operation
	Data    *Project `json:"data,omitempty"` // The project data (present only on success)
}

// ProjectsResponse represents an API response containing a list of projects.
type ProjectsResponse struct {
	Success bool      `json:"success"`        // Indicates if the operation was successful
	Message string    `json:"message"`        // Human-readable message about the operation
	Data    []Project `json:"data,omitempty"` // Array of project data
}

// ProjectType is a utility type for type-safe operations involving projects.
type ProjectType struct{}

// ProjectsTypeVal is a global instance of ProjectType for use in type-safe operations.
var ProjectsTypeVal = ProjectType{}
