package sdk

import "github.com/codearena-platform/codearena-platform/internal/iam/entity"

type AuthIdentityType = entity.AuthIdentityType

const (
	AuthIdentityTypeEmail = entity.AuthIdentityTypeEmail
	AuthIdentityTypePhone = entity.AuthIdentityTypePhone
)

type AuthMetadataType = entity.AuthMetadataType
type AuthIdentity = entity.AuthIdentity
