package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"

	"github.com/codearena-platform/codearena-platform/internal/iam/bootstrap"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/docs"
)

func main() {
	app := fiber.New(fiber.Config{
		ReadBufferSize: 8192,
	})

	cnf := bootstrap.SetupServer(app)

	for _, route := range app.GetRoutes() {
		if route.Method == "OPTIONS" || route.Method == "HEAD" || route.Method == "TRACE" || route.Method == "CONNECT" {
			continue
		}
		if route.Path == "/" {
			continue
		}
		log.Infof("%s %s", route.Method, route.Path)
	}

	err := docs.CreateOpenApiDoc("docs/iam.yaml")
	if err != nil {
		log.Fatal("failed to create OpenAPI doc: %w", err)
	}

	err = app.Listen(":" + cnf.Server.Port)
	if err != nil {
		log.Fatal(err)
		return
	}
}
