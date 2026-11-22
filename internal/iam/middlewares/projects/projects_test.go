package projects

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectService is a mock implementation of project.Service.
type MockProjectService struct {
	mock.Mock
}

func (m *MockProjectService) GetAll(ctx context.Context) ([]sdk.Project, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sdk.Project), args.Error(1)
}

func (m *MockProjectService) GetById(ctx context.Context, id string) (*sdk.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sdk.Project), args.Error(1)
}

func (m *MockProjectService) Create(ctx context.Context, project *sdk.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectService) Update(ctx context.Context, project *sdk.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectService) GetByName(ctx context.Context, name string) (*sdk.Project, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sdk.Project), args.Error(1)
}

func projectsFromResponse(t *testing.T, header string, user *sdk.User) []string {
	t.Helper()
	app := fiber.New()
	var captured []string
	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Context().SetUserValue(sdk.UserTypeVal, user)
		}
		return c.Next()
	})
	app.Get("/scope", NewMiddlewares(&MockProjectService{}).Projects, func(c *fiber.Ctx) error {
		value := c.Context().UserValue(sdk.ProjectsTypeVal)
		if projects, ok := value.([]string); ok {
			captured = projects
		}
		return c.SendStatus(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/scope", nil)
	if header != "" {
		req.Header.Set("X-Project-Ids", header)
	}
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	return captured
}

func TestProjects_HeaderNeverGrantsScope(t *testing.T) {
	for _, header := range []string{"project1", "project1,project2", "项目1", ",,"} {
		assert.Empty(t, projectsFromResponse(t, header, nil), "header %q must not grant a project scope", header)
	}
}

func TestProjects_ScopeComesFromVerifiedUser(t *testing.T) {
	user := &sdk.User{Id: "user-1", ProjectId: "project-a"}

	assert.Equal(t, []string{"project-a"}, projectsFromResponse(t, "", user))
	// A client-supplied header cannot add or replace the verified scope.
	assert.Equal(t, []string{"project-a"}, projectsFromResponse(t, "project-b", user))
	assert.Equal(t, []string{"project-a"}, projectsFromResponse(t, "project-a,project-b", user))
}

func TestProjects_AuthenticatedUserWithoutProjectHasNoScope(t *testing.T) {
	assert.Empty(t, projectsFromResponse(t, "project-a", &sdk.User{Id: "user-1"}))
}
