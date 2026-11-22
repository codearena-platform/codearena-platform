package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// migrations holds the versioned competition schema. Every migration is
// idempotent so an interrupted run can be retried safely; each migration is
// applied and recorded in the same transaction.
type migration struct {
	Version    int
	Name       string
	Statements []string
}

const migrationTable = "competition_schema_migrations"

var migrations = []migration{
	{
		Version: 1,
		Name:    "core schema",
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS bots (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				name TEXT NOT NULL,
				avatar_url TEXT,
				rating FLOAT DEFAULT 1200,
				matches INT DEFAULT 0,
				wins INT DEFAULT 0,
				created_at BIGINT
			)`,
			`CREATE TABLE IF NOT EXISTS scores (
				id SERIAL PRIMARY KEY,
				match_id TEXT NOT NULL,
				participant_id TEXT NOT NULL,
				points INT DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE TABLE IF NOT EXISTS tournaments (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				status TEXT NOT NULL,
				description TEXT,
				config JSONB,
				created_at BIGINT
			)`,
			`CREATE TABLE IF NOT EXISTS brackets (
				id TEXT PRIMARY KEY,
				tournament_id TEXT NOT NULL REFERENCES tournaments(id),
				level INT,
				match_id TEXT,
				position INT
			)`,
			`CREATE TABLE IF NOT EXISTS leagues (
				id TEXT PRIMARY KEY,
				tournament_id TEXT NOT NULL REFERENCES tournaments(id),
				season TEXT NOT NULL
			)`,
			`CREATE TABLE IF NOT EXISTS divisions (
				id TEXT PRIMARY KEY,
				league_id TEXT NOT NULL REFERENCES leagues(id),
				name TEXT NOT NULL,
				capacity INT,
				promotion_slots INT,
				relegation_slots INT
			)`,
			`CREATE TABLE IF NOT EXISTS standings (
				division_id TEXT NOT NULL REFERENCES divisions(id),
				bot_id TEXT NOT NULL REFERENCES bots(id),
				points INT DEFAULT 0,
				wins INT DEFAULT 0,
				draws INT DEFAULT 0,
				losses INT DEFAULT 0,
				matches INT DEFAULT 0,
				dano_causado FLOAT DEFAULT 0,
				PRIMARY KEY (division_id, bot_id)
			)`,
			`CREATE INDEX IF NOT EXISTS idx_scores_match_id ON scores(match_id)`,
			`CREATE INDEX IF NOT EXISTS idx_brackets_tournament_id ON brackets(tournament_id)`,
			`CREATE INDEX IF NOT EXISTS idx_leagues_tournament_id ON leagues(tournament_id)`,
			`CREATE INDEX IF NOT EXISTS idx_divisions_league_id ON divisions(league_id)`,
		},
	},
	{
		Version: 2,
		Name:    "hubs and tournament scoping",
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS hubs (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				visibility TEXT NOT NULL,
				owner_id TEXT NOT NULL,
				description TEXT,
				created_at BIGINT
			)`,
			`CREATE TABLE IF NOT EXISTS hub_memberships (
				hub_id TEXT NOT NULL REFERENCES hubs(id),
				user_id TEXT NOT NULL,
				role TEXT NOT NULL,
				joined_at BIGINT,
				PRIMARY KEY (hub_id, user_id)
			)`,
			`ALTER TABLE tournaments ADD COLUMN IF NOT EXISTS hub_id TEXT`,
			`CREATE INDEX IF NOT EXISTS idx_tournaments_hub_id ON tournaments(hub_id)`,
			`CREATE INDEX IF NOT EXISTS idx_hub_memberships_user_id ON hub_memberships(user_id)`,
		},
	},
}

// migrationStore separates the versioning logic from database/sql so the
// ordering and skip behavior can be tested without a live Postgres instance.
type migrationStore interface {
	AppliedVersions(ctx context.Context) (map[int]bool, error)
	Apply(ctx context.Context, m migration) error
}

type sqlMigrationStore struct {
	db *sql.DB
}

func (s sqlMigrationStore) AppliedVersions(ctx context.Context) (map[int]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT version FROM `+migrationTable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := map[int]bool{}
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func (s sqlMigrationStore) Apply(ctx context.Context, m migration) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, statement := range m.Statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migration %d (%s): %w", m.Version, m.Name, err)
		}
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO `+migrationTable+` (version, name) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING`,
		m.Version, m.Name); err != nil {
		return fmt.Errorf("record migration %d: %w", m.Version, err)
	}
	return tx.Commit()
}

// RunMigrations applies pending migrations in version order.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS `+migrationTable+` (
		version INT PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	return runMigrations(ctx, sqlMigrationStore{db: db})
}

func runMigrations(ctx context.Context, store migrationStore) error {
	applied, err := store.AppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}
		if err := store.Apply(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
