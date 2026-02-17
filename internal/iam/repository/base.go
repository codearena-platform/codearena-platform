package repository

import (
	"github.com/codearena-platform/codearena-platform/internal/iam/db"
)

// BaseRepository provides common database operations.
type BaseRepository interface {
	DB() db.DB
}

type baseRepository struct {
	db db.DB
}

func NewBaseRepository(db db.DB) BaseRepository {
	return &baseRepository{db: db}
}

func (r *baseRepository) DB() db.DB {
	return r.db
}
