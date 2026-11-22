package syncuser

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/constants"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/auth"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
	"github.com/gofiber/fiber/v2/log"
)

type service struct {
	authSvc auth.Service
}

func NewService(authSvc auth.Service) Service {
	return &service{
		authSvc: authSvc,
	}
}

func (s *service) HandleEvent(event utils.Event[sdk.User]) {
	if event == nil {
		return
	}
	if event.Name() == constants.EventUserUpdated {
		user := event.Payload()
		if user.Id == "" {
			return
		}
		err := s.authSvc.SynchronizeIdentity(event.Context(), user.Id)
		if err != nil {
			log.Errorw("failed to synchronize identity", "error", err, "userId", user.Id)
		}
	}
}
