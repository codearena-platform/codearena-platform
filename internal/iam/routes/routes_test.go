package routes

import (
	"os"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/config"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/cache"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/test"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/test/server"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutes(t *testing.T) {
	err := os.Setenv("JWT_SECRET", "abcd")
	require.NoError(t, err)
	cnf := config.NewAppConfig()

	app := fiber.New()

	d := test.SetupMockDB()
	cs := cache.NewMockService()
	svcs, err := server.GetServices(*cnf, cs, d)
	require.NoError(t, err)

	prv, err := server.InjectTestProviders(svcs, cs, d)
	require.NoError(t, err)

	RegisterRoutes(app, prv)

	// Check if routes are registered
	routes := app.GetRoutes()
	assert.NotEmpty(t, routes, "Routes should be registered")

	healthRouteFound := false
	userRouteFound := false

	for _, route := range routes {
		if route.Path == "/api/v1/iam/health" && route.Method == "GET" {
			healthRouteFound = true
		}
		if route.Path == "/api/v1/iam/user/v1/" && route.Method == "GET" {
			userRouteFound = true
		}
	}

	assert.True(t, healthRouteFound, "Health route should be registered at /api/v1/iam/health")
	assert.True(t, userRouteFound, "User route should be registered at /api/v1/iam/user/v1/")
}
