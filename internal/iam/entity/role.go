package entity

import (
	"errors"
	"time"
)

var ErrRoleNotFound = errors.New("role not found")
var ErrResourceNotFound = errors.New("resource not found")

type Role struct {
	Id          string               `json:"id"`
	ProjectId   string               `json:"project_id"`
	Description string               `json:"description"`
	Name        string               `json:"name"`
	Resources   map[string]Resources `json:"resources"`
	Enabled     bool                 `json:"enabled"`
	CreatedAt   *time.Time           `json:"created_at"`
	CreatedBy   string               `json:"created_by"`
	UpdatedAt   *time.Time           `json:"updated_at"`
	UpdatedBy   string               `json:"updated_by"`
}

type Resources struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Resource struct {
	ID          string     `json:"id"`
	ProjectId   string     `json:"project_id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   *time.Time `json:"created_at"`
	CreatedBy   string     `json:"created_by"`
	UpdatedAt   *time.Time `json:"updated_at"`
	UpdatedBy   string     `json:"updated_by"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
