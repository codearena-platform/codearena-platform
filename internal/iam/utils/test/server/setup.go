package server

import (
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/config"
	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares/auth"
	"github.com/codearena-platform/codearena-platform/internal/iam/middlewares/projects"
	"github.com/codearena-platform/codearena-platform/internal/iam/providers"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/cache"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/encrypt"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/jwt"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/httpclient"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/test/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/stretchr/testify/mock"
)

func GetServices(cnf config.AppConfig, cS cache.Service, d db.DB) (*providers.Service, error) {

	enc, err := encrypt.NewService(cnf.Encrypter.Key())
	if err != nil {
		return nil, fmt.Errorf("error creating encrypter: %w", err)
	}

	jwtSvc := jwt.NewService(cnf.Jwt.Secret())

	svcs := providers.NewServices(d, cS, enc, jwtSvc, cnf.Server.TokenCacheTTLInMinutes, cnf.Server.AuthProviderRefetchIntervalInMinutes)

	mockClientSvc := services.MockClientService{}
	mockProjectSvc := services.MockProjectService{}
	mockClientSvc.On("GetIAMClients", mock.Anything, mock.Anything).Return([]sdk.Client{}, nil)
	mockClientSvc.On("Subscribe", mock.Anything, mock.Anything).Return()
	mockProjectSvc.On("GetByName", mock.Anything, mock.Anything).Return(&sdk.Project{}, nil).Once()

	svcs.Clients = &mockClientSvc
	svcs.Projects = &mockProjectSvc

	return svcs, nil
}

func InjectTestProviders(svcs *providers.Service, cS cache.Service, d db.DB) (*providers.Provider, error) {

	pm := projects.NewMiddlewares(svcs.Projects)
	am, err := auth.NewMiddlewares(svcs.Auth, svcs.Clients)
	if err != nil {
		return nil, err
	}
	authClient, err := httpclient.GetIAMClient(svcs.Clients)
	if err != nil {
		return nil, err
	}

	pvd := &providers.Provider{
		S:          svcs,
		D:          d,
		C:          cS,
		PM:         pm,
		AM:         am,
		AuthClient: authClient,
	}

	// subscribe to client events for checking auth client
	svcs.Clients.Subscribe(constants.EventClientCreated, pvd)
	svcs.Clients.Subscribe(constants.EventClientUpdated, pvd)
	svcs.Clients.Subscribe(constants.EventClientCreated, svcs.Auth)
	svcs.Clients.Subscribe(constants.EventClientUpdated, svcs.Auth)

	// creating default project if it doesn't exist
	err = providers.CheckAndAddDefaultProject(svcs.Projects)
	if err != nil {
		log.Errorw("error checking and adding default project", "error", err)
		return nil, fmt.Errorf("error checking and adding default project: %w", err)
	}

	return pvd, nil
}

func SetupTestServer(app *fiber.App, cnf *config.AppConfig, svcs *providers.Service, cS cache.Service, db db.DB) *providers.Provider {
	prv, err := InjectTestProviders(svcs, cS, db)
	if err != nil {
		log.Fatalf("error injecting providers %s", err)
	}
	app.Use((*cnf).Handle)
	app.Use(providers.Handle(prv))
	app.Use(cors.New())

	app.Use(prv.PM.Projects)

	return prv
}
