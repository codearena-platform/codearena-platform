package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/codearena-platform/codearena-platform/internal/iam/config"
	"github.com/codearena-platform/codearena-platform/internal/iam/db"

	_ "github.com/codearena-platform/codearena-platform/internal/iam/db/migrations"
)

// NewDBConnection establishes a MongoDB connection and applies database migrations.
// This function initializes the database connection using the provided configuration
// and ensures the database schema is up to date by running all pending migrations.
//
// The function performs the following operations:
// 1. Creates MongoDB connection using the configured host
// 2. Applies all registered migrations to ensure schema consistency
// 3. Returns the database interface for use by services
//
// Parameters:
//   - cnf: Application configuration containing database settings
//
// Returns:
//   - db.DB: Database interface for service operations
//   - error: Error if connection fails or migrations cannot be applied
func NewDBConnection(cnf config.AppConfig) (db.DB, error) {
	conn, err := db.NewMongoConnection(cnf.DB.Host())
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}
	ctx := context.Background()
	// Explicit rollback takes precedence: operators can revert a bad release
	// without editing code.
	if version := os.Getenv("IAM_MIGRATION_ROLLBACK"); version != "" {
		if err := db.RollbackMigration(ctx, conn, version); err != nil {
			return nil, fmt.Errorf("error rolling back migration %s: %w", version, err)
		}
	}
	// IAM_MIGRATION_DRY_RUN=true inspects pending migrations without writes.
	if err := db.CheckAndRunMigrations(ctx, conn, db.WithDryRun(os.Getenv("IAM_MIGRATION_DRY_RUN") == "true")); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}
	return conn, nil
}
