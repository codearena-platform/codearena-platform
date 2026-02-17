package auth

import (
	"context"

	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/goiamuniverse"
)

func (s *service) HandleEvent(e utils.Event[sdk.Client]) {
	switch e.Name() {
	case goiamuniverse.EventClientCreated:
		s.handleClientUpdates(e)
	case goiamuniverse.EventClientUpdated:
		s.handleClientUpdates(e)
	default:
		return
	}

}

func (s *service) handleClientUpdates(e utils.Event[sdk.Client]) {
	s.cacheClientSecret(context.Background(), e.Payload().Id, e.Payload().Secret)
}
