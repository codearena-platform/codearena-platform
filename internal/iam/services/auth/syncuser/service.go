package syncuser

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils"
)

type Service interface {
	HandleEvent(event utils.Event[sdk.User])
}
