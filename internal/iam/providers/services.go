package providers

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/repository"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/auth"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/auth/syncuser"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/authprovider"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/cache"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/client"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/encrypt"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/jwt"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/policy"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/policy/system"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/project"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/resource"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/role"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/user"
	"github.com/codearena-platform/codearena-platform/internal/iam/usecase"
	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
)

// Service is a container that holds all business logic services for the Go IAM system.
// It provides centralized access to all domain services and manages their dependencies.
// Services are organized by domain and provide the core functionality for IAM operations.
type Service struct {
	Projects      project.Service      // Project management service
	Clients       client.Service       // OAuth2/OIDC client management service
	AuthProviders authprovider.Service // Authentication provider management service
	AuthSync      syncuser.Service     // User synchronization service
	Auth          auth.Service         // Authentication and token validation service
	Resources     resource.Service     // Resource management service
	User          user.Service         // User management and authorization service
	Role          role.Service         // Role-based access control service
	Policy        policy.Service       // Policy management service

	// UseCases
	AuthUC     usecase.AuthUseCase
	IdentityUC usecase.IdentityUseCase
	AccessUC   usecase.AccessUseCase
	AdminUC    usecase.AdminUseCase
}

// NewServices creates and configures all business logic services with their dependencies.
// This function initializes services in the correct order to satisfy dependencies and
// sets up event subscriptions for cross-service communication.
func NewServices(db db.DB, cache cache.Service, enc encrypt.Service, jwtSvc jwt.Service, tokenTTL int64, refetchTTL int64) *Service {
	// Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	authProviderRepo := repository.NewAuthProviderRepository(db, enc)
	clientRepo := repository.NewClientRepository(db)
	policyRepo := repository.NewPolicyRepository()

	psvc := project.NewService(projectRepo)
	rsvc := resource.NewService(resourceRepo)
	roleSvc := role.NewService(roleRepo)
	userSvc := user.NewService(userRepo, roleSvc)

	// subscribing to role updates
	roleSvc.Subscribe(constants.EventRoleUpdated, userSvc)
	// subscribing to resource create updates
	rsvc.Subscribe(constants.EventResourceCreated, system.NewAccessToCreatedResource(userSvc))
	rsvc.Subscribe(constants.EventResourceCreated, system.NewAddResourcesToUser(userSvc))
	rsvc.Subscribe(constants.EventResourceCreated, system.NewAddResourcesToRole(userSvc, roleSvc))
	// adding default policies to a user when gets created
	userSvc.Subscribe(constants.EventUserCreated, system.NewDefaultPoliciesOnUser(userSvc))

	apSvc := authprovider.NewService(authProviderRepo, psvc)
	csvc := client.NewService(clientRepo, psvc, apSvc, userSvc)

	authSvc := auth.NewService(apSvc, csvc, cache, jwtSvc, enc, userSvc, tokenTTL, refetchTTL)
	authSyncSvc := syncuser.NewService(authSvc)
	polSvc := policy.NewService(policyRepo)

	// UseCases
	authUC := usecase.NewAuthUseCase(authSvc, apSvc)
	identityUC := usecase.NewIdentityUseCase(userSvc)
	accessUC := usecase.NewAccessUseCase(roleSvc, rsvc, polSvc)
	adminUC := usecase.NewAdminUseCase(psvc, csvc)

	return &Service{
		Projects:      psvc,
		Clients:       csvc,
		AuthProviders: apSvc,
		Auth:          authSvc,
		User:          userSvc,
		Resources:     rsvc,
		Role:          roleSvc,
		Policy:        polSvc,
		AuthSync:      authSyncSvc,

		AuthUC:     authUC,
		IdentityUC: identityUC,
		AccessUC:   accessUC,
		AdminUC:    adminUC,
	}
}
