package jwt

import (
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	jwtp "github.com/golang-jwt/jwt/v4"
)

const (
	// tokenIssuer and tokenAudience are embedded in every token we issue and
	// checked on validation so tokens minted for another audience are rejected.
	tokenIssuer   = "codearena-iam"
	tokenAudience = "codearena"
)

type service struct {
	secret []byte
}

func NewService(secret sdk.MaskedBytes) Service {
	return &service{secret: []byte(secret)}
}

// hasAudience accepts both the single-string and array forms allowed for the
// registered "aud" claim.
func hasAudience(value interface{}, expected string) bool {
	switch aud := value.(type) {
	case string:
		return aud == expected
	case []string:
		return len(aud) == 1 && aud[0] == expected
	case []interface{}:
		if len(aud) != 1 {
			return false
		}
		item, ok := aud[0].(string)
		return ok && item == expected
	default:
		return false
	}
}

// GenerateToken generates a JWT token with the given claims
func (s service) GenerateToken(claims map[string]interface{}, expiryTimeInSeconds int64) (string, error) {
	if len(s.secret) == 0 {
		return "", fmt.Errorf("jwt secret is not configured")
	}
	jwtClaims := jwtp.MapClaims{}
	for k, v := range claims {
		// Callers must not be able to override the registered claims below.
		if k == "exp" || k == "iss" || k == "aud" {
			continue
		}
		jwtClaims[k] = v
	}
	jwtClaims["exp"] = expiryTimeInSeconds
	jwtClaims["iss"] = tokenIssuer
	jwtClaims["aud"] = tokenAudience
	token := jwtp.NewWithClaims(jwtp.SigningMethodHS256, jwtClaims)

	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("error signing jwt token with secret %w", err)
	}
	return tokenString, nil
}

// ValidateToken validates the given JWT token and returns the claims.
// Only tokens signed with HS256 for this issuer/audience are accepted.
func (s service) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwtp.Parse(tokenString, func(token *jwtp.Token) (interface{}, error) {
		// Pin the algorithm: accepting any HMAC method allows algorithm
		// confusion when the same secret is used with a different strength.
		if token.Method == nil || token.Method.Alg() != jwtp.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing jwt token %w", err)
	}
	claims, ok := token.Claims.(jwtp.MapClaims)
	if !ok || !token.Valid {
		// Never return empty claims with a nil error: callers treat that as a
		// successfully authenticated (but anonymous) identity.
		return nil, fmt.Errorf("error parsing jwt token: invalid claims")
	}
	if issuer, ok := claims["iss"].(string); !ok || issuer != tokenIssuer {
		return nil, fmt.Errorf("error parsing jwt token: invalid issuer")
	}
	if !hasAudience(claims["aud"], tokenAudience) {
		return nil, fmt.Errorf("error parsing jwt token: invalid audience")
	}

	result := map[string]interface{}{}
	for k, v := range claims {
		if k == "exp" || k == "iss" || k == "aud" {
			continue
		}
		result[k] = v
	}
	return result, nil
}
