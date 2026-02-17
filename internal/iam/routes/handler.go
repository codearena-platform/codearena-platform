package routes

import (
	"time"

	"github.com/codearena-platform/codearena-platform/internal/iam/providers"
)

type IAMHandler struct {
	prv       *providers.Provider
	startTime time.Time
}

func NewIAMHandler(prv *providers.Provider) *IAMHandler {
	return &IAMHandler{
		prv:       prv,
		startTime: time.Now(),
	}
}
