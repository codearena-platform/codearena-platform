package usecase

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/auth"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/authprovider"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type AuthUseCase interface {
	// Auth operations
	GetLoginUrl(ctx context.Context, clientId, authProviderId, state, redirectUrl, codeChallengeMethod, codeChallenge string) (string, error)
	Redirect(ctx context.Context, code, state string) (*sdk.AuthRedirectResponse, error)
	ClientCallback(ctx context.Context, code, codeChallenge, clientId, clietSecret string) (*sdk.AuthVerifyCodeResponse, error)
	GetIdentity(ctx context.Context, accessToken string, forceFetch bool) (*entity.User, error)
	SynchronizeIdentity(ctx context.Context, userId string) error
	ClientCredentials(ctx context.Context, clientId, clientSecret string) (*sdk.AuthVerifyCodeResponse, error)

	// AuthProvider operations
	CreateAuthProvider(ctx context.Context, provider *entity.AuthProvider) error
	UpdateAuthProvider(ctx context.Context, provider *entity.AuthProvider) error
	GetAuthProvider(ctx context.Context, id string, dontCheckProjects bool) (*entity.AuthProvider, error)
	GetAllAuthProviders(ctx context.Context, queryParams sdk.AuthProviderQueryParams) ([]entity.AuthProvider, error)

	// Events
	HandleClientEvent(event utils.Event[entity.Client])
}

// NOTE: The above HandleClientEvent has a small syntax error in the package name if I just copy-paste.
// I'll fix the signature in the actual file.

type authUseCase struct {
	authSvc         auth.Service
	authProviderSvc authprovider.Service
}

func NewAuthUseCase(authSvc auth.Service, authProviderSvc authprovider.Service) AuthUseCase {
	return &authUseCase{
		authSvc:         authSvc,
		authProviderSvc: authProviderSvc,
	}
}

func (u *authUseCase) GetLoginUrl(ctx context.Context, clientId, authProviderId, state, redirectUrl, codeChallengeMethod, codeChallenge string) (string, error) {
	return u.authSvc.GetLoginUrl(ctx, clientId, authProviderId, state, redirectUrl, codeChallengeMethod, codeChallenge)
}

func (u *authUseCase) Redirect(ctx context.Context, code, state string) (*sdk.AuthRedirectResponse, error) {
	return u.authSvc.Redirect(ctx, code, state)
}

func (u *authUseCase) ClientCallback(ctx context.Context, code, codeChallenge, clientId, clietSecret string) (*sdk.AuthVerifyCodeResponse, error) {
	return u.authSvc.ClientCallback(ctx, code, codeChallenge, clientId, clietSecret)
}

func (u *authUseCase) GetIdentity(ctx context.Context, accessToken string, forceFetch bool) (*entity.User, error) {
	return u.authSvc.GetIdentity(ctx, accessToken, forceFetch)
}

func (u *authUseCase) SynchronizeIdentity(ctx context.Context, userId string) error {
	return u.authSvc.SynchronizeIdentity(ctx, userId)
}

func (u *authUseCase) ClientCredentials(ctx context.Context, clientId, clientSecret string) (*sdk.AuthVerifyCodeResponse, error) {
	return u.authSvc.ClientCredentials(ctx, clientId, clientSecret)
}

func (u *authUseCase) CreateAuthProvider(ctx context.Context, provider *entity.AuthProvider) error {
	return u.authProviderSvc.Create(ctx, provider)
}

func (u *authUseCase) UpdateAuthProvider(ctx context.Context, provider *entity.AuthProvider) error {
	return u.authProviderSvc.Update(ctx, provider)
}

func (u *authUseCase) GetAuthProvider(ctx context.Context, id string, dontCheckProjects bool) (*entity.AuthProvider, error) {
	return u.authProviderSvc.Get(ctx, id, dontCheckProjects)
}

func (u *authUseCase) GetAllAuthProviders(ctx context.Context, queryParams sdk.AuthProviderQueryParams) ([]entity.AuthProvider, error) {
	return u.authProviderSvc.GetAll(ctx, queryParams)
}

func (u *authUseCase) HandleClientEvent(event utils.Event[entity.Client]) {
	u.authSvc.HandleEvent(event)
}
