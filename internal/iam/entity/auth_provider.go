package entity

import (
	"context"
	"errors"
	"time"
)

var ErrAuthProviderNotFound = errors.New("auth provider not found")

type AuthToken struct {
	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token"`
	ExpiresAt            time.Time `json:"expires_at"`
	AuthProviderID       string    `json:"auth_provider_id"`
	CodeChallengeMethod  string    `json:"code_challenge_method"`
	CodeChallenge        string    `json:"code_challenge"`
	ClientId             string    `json:"client_id"`
	ServiceAccountUserId string    `json:"service_account_user_id"`
}

type AuthProviderType string

const (
	AuthProviderTypeGoogle    AuthProviderType = "GOOGLE"
	AuthProviderTypeMicrosoft AuthProviderType = "MICROSOFT"
	AuthProviderTypeGitHub    AuthProviderType = "GITHUB"
	AuthProviderTypeOIDC      AuthProviderType = "OIDC"
)

type AuthProvider struct {
	Id        string              `json:"id"`
	Name      string              `json:"name"`
	Icon      string              `json:"icon"`
	Provider  AuthProviderType    `json:"provider"`
	Params    []AuthProviderParam `json:"params"`
	ProjectId string              `json:"project_id"`
	Enabled   bool                `json:"enabled"`
	CreatedAt *time.Time          `json:"created_at"`
	UpdatedAt *time.Time          `json:"updated_at"`
	CreatedBy string              `json:"created_by"`
	UpdatedBy string              `json:"updated_by"`
}

func (a AuthProvider) GetParam(key string) string {
	for _, p := range a.Params {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

type AuthProviderParam struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	Key      string `json:"key"`
	IsSecret bool   `json:"is_secret"`
}

type ServiceProvider interface {
	GetAuthCodeUrl(state string) string
	VerifyCode(ctx context.Context, code string) (*AuthToken, error)
	RefreshToken(refreshToken string) (*AuthToken, error)
	GetIdentity(token string) ([]AuthIdentity, error)
	HasRefreshTokenFlow() bool
}
