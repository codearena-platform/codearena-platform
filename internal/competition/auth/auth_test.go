package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamSDK "github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type stubVerifier struct {
	identity *Identity
	err      error
}

func (s stubVerifier) Verify(_ context.Context, token string) (*Identity, error) {
	if token != "valid-token" {
		return nil, errors.New("invalid token")
	}
	return s.identity, s.err
}

func TestRequireUser_IdentityComesFromToken(t *testing.T) {
	app := fiber.New()
	app.Post("/bots", RequireUser(stubVerifier{identity: &Identity{ID: "token-user", Email: "user@example.com"}}), func(c *fiber.Ctx) error {
		identity, err := IdentityFromFiber(c)
		if err != nil {
			return c.Status(http.StatusUnauthorized).SendString(err.Error())
		}
		return c.SendString(identity.ID)
	})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantBody   string
	}{
		{name: "no token", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", authHeader: "Bearer bad-token", wantStatus: http.StatusUnauthorized},
		{name: "valid token", authHeader: "Bearer valid-token", wantStatus: http.StatusOK, wantBody: "token-user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/bots", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			// A forged identity header must never influence the result.
			req.Header.Set("X-User-ID", "attacker")
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantBody != "" {
				buf := make([]byte, 64)
				n, _ := resp.Body.Read(buf)
				assert.Equal(t, tt.wantBody, string(buf[:n]))
			}
		})
	}
}

func TestIAMVerifier(t *testing.T) {
	t.Run("successful introspection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/iam/me", r.URL.Path)
			assert.Equal(t, "Bearer valid-token", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode(iamSDK.UserResponse{
				Success: true,
				Data:    &iamSDK.User{Id: "user-1", Email: "user@example.com", ProjectId: "project-1"},
			})
		}))
		defer server.Close()

		identity, err := NewIAMVerifier(server.URL).Verify(context.Background(), "valid-token")
		require.NoError(t, err)
		assert.Equal(t, "user-1", identity.ID)
		assert.Equal(t, "user@example.com", identity.Email)
		assert.Equal(t, "project-1", identity.ProjectID)
	})

	t.Run("IAM rejects token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		identity, err := NewIAMVerifier(server.URL).Verify(context.Background(), "bad-token")
		assert.Error(t, err)
		assert.Nil(t, identity)
	})

	t.Run("missing user in response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(iamSDK.UserResponse{Success: true})
		}))
		defer server.Close()

		identity, err := NewIAMVerifier(server.URL).Verify(context.Background(), "valid-token")
		assert.Error(t, err)
		assert.Nil(t, identity)
	})
}

func TestUnaryInterceptor(t *testing.T) {
	interceptor := UnaryInterceptor(stubVerifier{identity: &Identity{ID: "grpc-user"}})
	handler := func(ctx context.Context, _ interface{}) (interface{}, error) {
		identity, err := IdentityFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return identity.ID, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/competition.MatchService/RegisterBot"}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer valid-token"))
	result, err := interceptor(ctx, nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "grpc-user", result)

	// req.UserId and other payload fields are ignored by the interceptor.
	unauthenticated, err := interceptor(context.Background(), nil, info, handler)
	assert.Nil(t, unauthenticated)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthorizeMatchResult(t *testing.T) {
	t.Setenv("MATCH_RESULT_SERVICE_EMAIL", "judge@codearena.local")
	assert.NoError(t, AuthorizeMatchResult(&Identity{Email: "judge@codearena.local"}))
	assert.Error(t, AuthorizeMatchResult(&Identity{Email: "attacker@example.com"}))
}
