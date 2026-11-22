package postgres

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

func InitPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := RunMigrations(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
