package postgres

import (
	"database/sql"
)

type PGUserRepository struct {
	db *sql.DB
}

func NewPGUserRepository(db *sql.DB) *PGUserRepository {
	return &PGUserRepository{db: db}
}

// Add methods as per interface if needed, or keep minimal for now
