package system

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/user"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type defaultPoliciesOnUser struct {
	id      string
	userSvc user.Service
}

func NewDefaultPoliciesOnUser(userSvc user.Service) defaultPoliciesOnUser {
	return defaultPoliciesOnUser{id: "@policy/system/default_policies_on_user", userSvc: userSvc}
}

func (a defaultPoliciesOnUser) ID() string {
	return a.id
}

func (a defaultPoliciesOnUser) HandleEvent(event utils.Event[sdk.User]) {
	log.Debugw("received user event", "event", event.Name())
	userId := event.Payload().Id
	pl := NewAccessToCreatedResource(nil)
	err := a.userSvc.AddPolicyToUser(event.Context(), userId, map[string]sdk.UserPolicy{pl.ID(): {Name: pl.Name()}})
	if err != nil {
		log.Errorw("error updating user with default policies while handling user create event", "userId", userId, "error", err)
		return
	}
	log.Infow("successfully added default policies to user", "userId", userId)
}
