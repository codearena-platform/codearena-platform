package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func InitPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Simple migration for demo
	query := `
	CREATE TABLE IF NOT EXISTS bots (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		avatar_url TEXT,
		rating FLOAT DEFAULT 1200,
		matches INT DEFAULT 0,
		wins INT DEFAULT 0,
		created_at BIGINT
	);

	CREATE TABLE IF NOT EXISTS scores (
		id SERIAL PRIMARY KEY,
		match_id TEXT NOT NULL,
		participant_id TEXT NOT NULL,
		points INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tournaments (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		description TEXT,
		config JSONB,
		created_at BIGINT
	);

	CREATE TABLE IF NOT EXISTS brackets (
		id TEXT PRIMARY KEY,
		tournament_id TEXT NOT NULL REFERENCES tournaments(id),
		level INT,
		match_id TEXT,
		position INT
	);

	CREATE TABLE IF NOT EXISTS leagues (
		id TEXT PRIMARY KEY,
		tournament_id TEXT NOT NULL REFERENCES tournaments(id),
		season TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS divisions (
		id TEXT PRIMARY KEY,
		league_id TEXT NOT NULL REFERENCES leagues(id),
		name TEXT NOT NULL,
		capacity INT,
		promotion_slots INT,
		relegation_slots INT
	);

	CREATE TABLE IF NOT EXISTS standings (
		division_id TEXT NOT NULL REFERENCES divisions(id),
		bot_id TEXT NOT NULL REFERENCES bots(id),
		points INT DEFAULT 0,
		wins INT DEFAULT 0,
		draws INT DEFAULT 0,
		losses INT DEFAULT 0,
		matches INT DEFAULT 0,
		dano_causado FLOAT DEFAULT 0,
		PRIMARY KEY (division_id, bot_id)
	);
	`
	_, err = db.Exec(query)
	return db, err
}
