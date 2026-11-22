package policy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func testApp(user *sdk.User, projects []string) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Context().SetUserValue(sdk.UserTypeVal, user)
		}
		if projects != nil {
			c.Context().SetUserValue(sdk.ProjectsTypeVal, projects)
		}
		return c.Next()
	})
	app.Use(Enforce)
	handler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	app.Get("/resource", handler)
	app.Post("/resource", handler)
	return app
}

func TestEnforce_InternalCallsPass(t *testing.T) {
	app := testApp(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/resource", strings.NewReader(`{"project_id":"project-b"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEnforce_RequiresVerifiedProjectScope(t *testing.T) {
	app := testApp(&sdk.User{Id: "user-1"}, []string{})
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/resource", nil))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestEnforce_RejectsCrossTenantBody(t *testing.T) {
	app := testApp(&sdk.User{Id: "user-1"}, []string{"project-a"})
	for _, body := range []string{
		`{"project_id":"project-b"}`,
		`{"projectId":"project-b"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/resource", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode, "body %s", body)
	}
}

func TestEnforce_RejectsCrossTenantQuery(t *testing.T) {
	app := testApp(&sdk.User{Id: "user-1"}, []string{"project-a"})
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/resource?project_id=project-b", nil))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestEnforce_AllowsInScopeRequests(t *testing.T) {
	app := testApp(&sdk.User{Id: "user-1"}, []string{"project-a"})
	req := httptest.NewRequest(http.MethodPost, "/resource", strings.NewReader(`{"project_id":"project-a"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
