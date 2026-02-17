package entity

import (
	"errors"
	"time"
)

var ErrClientNotFound = errors.New("client not found")

type Client struct {
	Id                    string     `json:"id"`
	Name                  string     `json:"name"`
	Description           string     `json:"description"`
	Secret                string     `json:"secret"`
	Tags                  []string   `json:"tags"`
	RedirectURLs          []string   `json:"redirect_urls"`
	Scopes                []string   `json:"scopes"`
	ProjectId             string     `json:"project_id"`
	DefaultAuthProviderId string     `json:"default_auth_provider_id"`
	IAMClient             bool       `json:"iam_client"`
	LinkedUserId          string     `json:"linked_user_id"`
	ServiceAccountEmail   string     `json:"service_account_email"`
	Enabled               bool       `json:"enabled"`
	CreatedAt             *time.Time `json:"created_at"`
	CreatedBy             string     `json:"created_by"`
	UpdatedAt             *time.Time `json:"updated_at"`
	UpdatedBy             string     `json:"updated_by"`
}

func (c Client) IsServiceAccount() bool {
	return c.HasIAMAuthProvider() && c.LinkedUserId != ""
}

func (c Client) HasIAMAuthProvider() bool {
	return c.DefaultAuthProviderId == ""
}
