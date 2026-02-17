package routes

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/gofiber/fiber/v2"
)

var authTags = []string{"Auth"}

func (h *IAMHandler) RegisterAuthRoutes(router fiber.Router) {
	authGroup := router.Group("/auth")

	authGroup.Get("/login", h.Login)
	authGroup.Get("/authp-callback", h.AuthCallback)
	authGroup.Get("/verify", h.VerifyToken)
	authGroup.Post("/client", h.ClientCredentials)
}

func (h *IAMHandler) Login(c *fiber.Ctx) error {
	codeChallengeMethod := c.Query("code_challenge_method", "")
	if len(codeChallengeMethod) != 0 && strings.Compare(codeChallengeMethod, "S256") != 0 {
		return sdk.AuthProviderBadRequest("invalid code challenge. Only S256 is supported", c)
	}
	url, err := h.prv.S.AuthUC.GetLoginUrl(c.Context(), c.Query("client_id", ""), c.Query("auth_provider", ""), c.Query("state", ""), c.Query("redirect_url", ""), c.Query("code_challenge_method", ""), c.Query("code_challenge", ""))
	if err != nil {
		return sdk.AuthProviderInternalServerError(err.Error(), c)
	}
	if c.Query("postback") == "true" {
		return c.JSON(sdk.AuthLoginResponse{
			Success: true,
			Message: "Login URL generated successfully",
			Data:    sdk.AuthLoginDataResponse{LoginUrl: url},
		})
	}
	return c.Redirect(url, http.StatusTemporaryRedirect)
}

func (h *IAMHandler) AuthCallback(c *fiber.Ctx) error {
	resp, err := h.prv.S.AuthUC.Redirect(c.Context(), c.Query("code"), c.Query("state"))
	if err != nil {
		return sdk.AuthProviderInternalServerError(err.Error(), c)
	}
	if c.Query("postback") == "true" {
		return c.JSON(sdk.AuthRedirectResponse{RedirectUrl: resp.RedirectUrl})
	}
	return c.Redirect(resp.RedirectUrl, http.StatusTemporaryRedirect)
}

func (h *IAMHandler) VerifyToken(c *fiber.Ctx) error {
	code := c.Query("code")
	codeChallenge := c.Query("code_challenge")
	clientId := c.Query("client_id")
	var clientSecret string

	if len(codeChallenge) == 0 || len(clientId) == 0 {
		clId, clSec, ok := h.getClientDetails(c)
		if !ok {
			return sdk.AuthProviderBadRequest("missing or invalid authorization header", c)
		}
		clientId = clId
		clientSecret = clSec
	}
	resp, err := h.prv.S.AuthUC.ClientCallback(c.Context(), code, codeChallenge, clientId, clientSecret)
	if err != nil {
		return sdk.AuthProviderInternalServerError(err.Error(), c)
	}
	return c.JSON(sdk.AuthCallbackResponse{
		Success: true,
		Message: "Callback successful",
		Data:    resp,
	})
}

func (h *IAMHandler) ClientCredentials(c *fiber.Ctx) error {
	payload := new(sdk.ClientCredentialsRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(sdk.ClientCredentialsResponse{Success: false, Message: err.Error()})
	}
	resp, err := h.prv.S.AuthUC.ClientCredentials(c.Context(), payload.ClientId, payload.ClientSecret)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(sdk.ClientCredentialsResponse{Success: false, Message: err.Error()})
	}
	return c.JSON(sdk.ClientCredentialsResponse{
		Success: true,
		Message: "Authentication successful",
		Data:    resp,
	})
}

func (h *IAMHandler) getClientDetails(c *fiber.Ctx) (string, string, bool) {
	headers := c.GetReqHeaders()
	authHeaders := headers["Authorization"]
	if len(authHeaders) == 0 {
		return "", "", false
	}
	parts := strings.SplitN(authHeaders[0], " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", false
	}
	credentials, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}
	creds := strings.SplitN(string(credentials), ":", 2)
	if len(creds) != 2 {
		return "", "", false
	}
	return creds[0], creds[1], true
}
