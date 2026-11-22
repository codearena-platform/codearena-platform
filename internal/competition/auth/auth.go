// Package auth provides token verification and identity propagation for the
// competition service. Identities are derived exclusively from a verified IAM
// token; request headers and payload fields are never trusted.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	iamSDK "github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Identity is the authenticated principal extracted from an IAM token.
type Identity struct {
	ID        string
	Email     string
	ProjectID string
}

// Verifier validates a bearer token and returns the identity it belongs to.
type Verifier interface {
	Verify(ctx context.Context, token string) (*Identity, error)
}

type iamVerifier struct {
	client *http.Client
	meURL  string
}

// NewIAMVerifier returns a verifier that introspects tokens through IAM's
// authenticated /me endpoint. It fails closed when IAM is unreachable.
func NewIAMVerifier(baseURL string) Verifier {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://iam:7000"
	}
	return &iamVerifier{
		client: &http.Client{Timeout: 5 * time.Second},
		meURL:  baseURL + "/api/v1/iam/me",
	}
}

func (v *iamVerifier) Verify(ctx context.Context, token string) (*Identity, error) {
	if token == "" {
		return nil, errors.New("empty bearer token")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.meURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build IAM identity request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("IAM identity request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IAM rejected the token (status %d)", resp.StatusCode)
	}

	var payload iamSDK.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode IAM identity response: %w", err)
	}
	if !payload.Success || payload.Data == nil || payload.Data.Id == "" {
		return nil, errors.New("IAM response did not contain a valid identity")
	}
	return &Identity{
		ID:        payload.Data.Id,
		Email:     payload.Data.Email,
		ProjectID: payload.Data.ProjectId,
	}, nil
}

type identityKey struct{}

// AllowInsecure is the explicit local-development escape hatch. It mirrors the
// IAM middleware flag and must never be enabled in a deployed environment.
func AllowInsecure() bool {
	return os.Getenv("IAM_ALLOW_INSECURE") == "true"
}

func insecureIdentity() *Identity {
	log.Print("WARNING: IAM_ALLOW_INSECURE=true; competition is using a synthetic development identity")
	return &Identity{ID: "local-dev", Email: "local-dev@localhost"}
}

func bearerToken(header string) (string, bool) {
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	return token, token != ""
}

// RequireUser rejects unauthenticated HTTP requests and stores the verified
// identity in the Fiber locals for handlers to consume.
func RequireUser(v Verifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := bearerToken(c.Get("Authorization"))
		if !ok {
			if AllowInsecure() {
				c.Locals(identityKey{}, insecureIdentity())
				return c.Next()
			}
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
		}
		identity, err := v.Verify(c.Context(), token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals(identityKey{}, identity)
		return c.Next()
	}
}

// IdentityFromFiber returns the identity set by RequireUser.
func IdentityFromFiber(c *fiber.Ctx) (*Identity, error) {
	identity, ok := c.Locals(identityKey{}).(*Identity)
	if !ok || identity == nil || identity.ID == "" {
		return nil, errors.New("authenticated identity missing")
	}
	return identity, nil
}

// UnaryInterceptor performs the same verification for gRPC calls, reading the
// token from the standard authorization metadata key.
func UnaryInterceptor(v Verifier) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("authorization"); len(values) > 0 {
				if parsed, ok := bearerToken(values[0]); ok {
					token = parsed
				}
			}
		}
		if token == "" {
			if AllowInsecure() {
				ctx = context.WithValue(ctx, identityKey{}, insecureIdentity())
				return handler(ctx, req)
			}
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		identity, err := v.Verify(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		ctx = context.WithValue(ctx, identityKey{}, identity)
		return handler(ctx, req)
	}
}

// IdentityFromContext returns the identity set by UnaryInterceptor.
func IdentityFromContext(ctx context.Context) (*Identity, error) {
	identity, ok := ctx.Value(identityKey{}).(*Identity)
	if !ok || identity == nil || identity.ID == "" {
		return nil, status.Error(codes.Unauthenticated, "authenticated identity missing")
	}
	return identity, nil
}

// AuthorizeMatchResult constrains result callbacks to a configured service
// identity. Without the configuration any verified token is accepted, which is
// still strictly stronger than the previous unauthenticated endpoint.
func AuthorizeMatchResult(identity *Identity) error {
	expected := os.Getenv("MATCH_RESULT_SERVICE_EMAIL")
	if expected == "" {
		log.Print("WARNING: MATCH_RESULT_SERVICE_EMAIL is not set; any authenticated identity may report match results")
		return nil
	}
	if identity.Email != expected {
		return errors.New("caller is not allowed to report match results")
	}
	return nil
}
