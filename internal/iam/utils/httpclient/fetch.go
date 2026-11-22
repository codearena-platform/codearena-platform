package httpclient

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/client"
	"github.com/gofiber/fiber/v2/log"
)

func GetIAMClient(svc client.Service) (*sdk.Client, error) {
	prvs, err := svc.GetIAMClients(context.Background(), sdk.ClientQueryParams{
		IAMClient: true,
	})
	if err != nil {
		return nil, err
	}
	if len(prvs) == 0 {
		log.Warn("No IAM client found: authenticated routes will reject requests (set IAM_ALLOW_INSECURE=true only for local development)")
		return nil, nil
	}
	return &prvs[0], nil
}
