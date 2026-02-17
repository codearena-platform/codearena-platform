package bootstrap

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/config"
	"github.com/codearena-platform/codearena-platform/internal/iam/providers"
	"github.com/codearena-platform/codearena-platform/internal/iam/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupServer(app *fiber.App) *config.AppConfig {
	cnf := config.NewAppConfig()
	log.Infow("Loaded Configurations",
		"host", cnf.Server.Host,
		"port", cnf.Server.Port,
		"env", cnf.Deployment.Environment,
		"app_name", cnf.Deployment.Name,
	)
	prv, err := providers.InjectDefaultProviders(*cnf)
	if err != nil {
		log.Fatalf("error injecting providers %s", err)
	}
	app.Use((*cnf).Handle)
	app.Use(providers.Handle(prv))
	app.Use(cors.New())

	app.Use(prv.PM.Projects)
	routes.RegisterRoutes(app, prv)

	return cnf
}
