package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2/log"
)

// Cached client secrets are credentials: keep them only long enough to avoid
// hammering the database during a burst of token requests.
const clientSecretCacheTTL = 5 * time.Minute

func secretCacheKey(clientID string) string {
	return fmt.Sprintf("client-%s", clientID)
}

func (s *service) cacheClientSecret(ctx context.Context, clientId string, secret string) {
	err := s.cacheSvc.Set(ctx, secretCacheKey(clientId), secret, clientSecretCacheTTL)
	if err != nil {
		log.Errorf("failed to cache client secret: %w", err)
	}
}

func (s *service) getClientSecret(ctx context.Context, clientId string) (string, error) {
	secret, err := s.cacheSvc.Get(ctx, secretCacheKey(clientId))
	if err == nil {
		return secret, nil
	}
	cl, err := s.clientSvc.Get(ctx, clientId, true)
	if err != nil {
		return "", fmt.Errorf("couldn't get the client even from db: %w", err)
	}
	err = s.cacheSvc.Set(ctx, secretCacheKey(clientId), cl.Secret, clientSecretCacheTTL)
	if err != nil {
		log.Errorf("failed to cache client secret: %w", err)
	}
	return cl.Secret, nil
}

func (s *service) invalidateClientSecret(ctx context.Context, clientId string) {
	if err := s.cacheSvc.Delete(ctx, secretCacheKey(clientId)); err != nil {
		log.Errorf("failed to invalidate client secret cache: %w", err)
	}
}

func (s *service) handlePrivateClient(ctx context.Context, clientId, clientSecret string) error {
	secret, err := s.getClientSecret(ctx, clientId)
	if err != nil {
		return fmt.Errorf("error getting client secret: %w", err)
	}
	err = s.clientSvc.VerifySecret(clientSecret, secret)
	if err != nil {
		return fmt.Errorf("invalid client secret: %w", err)
	}
	return nil
}

// requireClientBinding verifies that the authorization code was issued to the
// client that is redeeming it. Without this check a client can steal another
// client's code.
func requireClientBinding(clientID string, token sdk.AuthToken) error {
	if clientID == "" || token.ClientId == "" {
		return errors.New("client id is required")
	}
	if subtle.ConstantTimeCompare([]byte(clientID), []byte(token.ClientId)) != 1 {
		return errors.New("invalid client id")
	}
	return nil
}

// handlePublicClient implements RFC 7636 S256. The value stored on the token
// is the challenge (BASE64URL(SHA256(verifier))) created at authorize time;
// the value supplied here is the verifier.
func (s *service) handlePublicClient(clientId, codeVerifier string, token sdk.AuthToken) error {
	if token.CodeChallengeMethod != "S256" {
		return errors.New("invalid code challenge method")
	}
	if codeVerifier == "" || token.CodeChallenge == "" {
		return errors.New("invalid code verifier")
	}
	challenge := generateCodeChallengeS256(codeVerifier)
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(token.CodeChallenge)) != 1 {
		log.Debugw("invalid code verifier", "calculated_challenge", challenge, "expected_challenge", token.CodeChallenge)
		return errors.New("invalid code verifier")
	}
	if err := requireClientBinding(clientId, token); err != nil {
		return err
	}
	return nil
}

func generateCodeChallengeS256(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
