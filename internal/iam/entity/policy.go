package entity

import (
	"errors"

	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
)

var ErrPolicyNotFound = errors.New("policy not found")

type Policy struct {
	Id          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Definition  PolicyDefinition `json:"definition"`
}

type PolicyDefinition struct {
	Arguments []PolicyArgument `json:"arguments,omitempty"`
}

type PolicyArgument struct {
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	DataType    constants.DataType `json:"data_type,omitempty"`
}
