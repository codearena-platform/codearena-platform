package repository

import "github.com/codearena-platform/codearena-platform/internal/competition/services/entity"

type BotRepository interface {
	Save(bot *entity.Bot) error
	GetByID(id string) (*entity.Bot, error)
	GetByUserID(userID string) ([]*entity.Bot, error)
	UpdateRating(botID string, newRating float64) error
}

type BotVersionRepository interface {
	Save(version *entity.BotVersion) error
	GetLatest(botID string) (*entity.BotVersion, error)
	GetVersions(botID string) ([]*entity.BotVersion, error)
}
