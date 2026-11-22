package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MigrationFunc represents a function that performs a database migration.
// It receives a context and database interface to execute migration operations.
type MigrationFunc func(ctx context.Context, db DB) error

// MigrationInfo contains metadata and implementation for a database migration.
// Each migration should have a unique version identifier and both up/down functions.
type MigrationInfo struct {
	Version     string        // Unique version identifier for the migration
	Name        string        // Human-readable name of the migration
	Description string        // Detailed description of what the migration does
	Up          MigrationFunc // Function to apply the migration
	Down        MigrationFunc // Function to rollback the migration
}

var registeredMigrations []MigrationInfo

// RegisterMigration registers a new migration to be tracked and executed.
// Migrations should be registered in the order they should be applied.
func RegisterMigration(migration MigrationInfo) {
	registeredMigrations = append(registeredMigrations, migration)
}

// MigrationOptions controls how migrations are executed.
type MigrationOptions struct {
	// DryRun inspects pending migrations and checks checksums without applying
	// any changes or writing migration records.
	DryRun bool
}

// MigrationOption customizes migration execution.
type MigrationOption func(*MigrationOptions)

// WithDryRun enables or disables dry-run mode.
func WithDryRun(dryRun bool) MigrationOption {
	return func(o *MigrationOptions) { o.DryRun = dryRun }
}

// CheckAndRunMigrations applies pending migrations in registration order.
// Apply and record happen in a single transaction when the backend supports it.
func CheckAndRunMigrations(ctx context.Context, conn DB, opts ...MigrationOption) error {
	options := MigrationOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	log.Info("Checking database migrations...")
	if options.DryRun {
		log.Warn("Running migrations in dry-run mode: no changes will be applied")
	}

	for _, migration := range registeredMigrations {
		if err := runMigration(ctx, conn, migration, options); err != nil {
			return err
		}
	}

	log.Info("All migrations checked and applied successfully")
	return nil
}

func migrationChecksum(migration MigrationInfo) string {
	h := sha256.New()
	parts := []string{
		migration.Version,
		migration.Name,
		migration.Description,
		functionName(migration.Up),
		functionName(migration.Down),
	}
	// NUL separators keep the hash unambiguous if a field contains a slash or
	// another part of a function name.
	_, _ = h.Write([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(h.Sum(nil))
}

func functionName(fn interface{}) string {
	if fn == nil {
		return ""
	}
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func || v.IsNil() {
		return ""
	}
	if f := runtime.FuncForPC(v.Pointer()); f != nil {
		return f.Name()
	}
	return ""
}

func runMigration(ctx context.Context, conn DB, migration MigrationInfo, options MigrationOptions) error {
	model := models.GetMigrationModel()
	filter := bson.M{model.VersionKey: migration.Version}

	var existing models.Migration
	err := conn.FindOne(ctx, model, filter).Decode(&existing)
	checksum := migrationChecksum(migration)

	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		if options.DryRun {
			log.Infof("DRY RUN: would apply migration %s: %s", migration.Version, migration.Name)
			return nil
		}
		log.Infof("Applying migration %s: %s", migration.Version, migration.Name)
		apply := func(txCtx context.Context) error {
			if err := migration.Up(txCtx, conn); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
			}
			now := time.Now()
			record := models.Migration{
				Id:          migration.Version,
				Version:     migration.Version,
				Name:        migration.Name,
				Description: migration.Description,
				AppliedAt:   &now,
				Checksum:    checksum,
				CreatedAt:   &now,
				CreatedBy:   "system",
				UpdatedAt:   &now,
				UpdatedBy:   "system",
			}
			if _, err := conn.InsertOne(txCtx, model, record); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}
			return nil
		}
		if tx, ok := conn.(TransactionalDB); ok {
			if err := tx.WithTransaction(ctx, apply); err != nil {
				return err
			}
		} else {
			log.Warnf("Database does not support transactions; migration %s and its record are not atomic", migration.Version)
			if err := apply(ctx); err != nil {
				return err
			}
		}
		log.Infof("Successfully applied migration %s", migration.Version)
		return nil
	case err != nil:
		return fmt.Errorf("failed to check migration %s: %w", migration.Version, err)
	default:
		if existing.Checksum == "" {
			// Records written by older versions predate checksums. Accept them
			// so upgrading does not fail, but make the gap visible.
			log.Warnf("Migration %s has no checksum (applied by an older version)", migration.Version)
			return nil
		}
		if existing.Checksum != checksum {
			return fmt.Errorf("migration %s checksum mismatch: recorded %s, computed %s", migration.Version, existing.Checksum, checksum)
		}
		log.Debugf("Migration %s already applied", migration.Version)
		return nil
	}
}

// RollbackMigration reverts an applied migration and removes its record in one
// transaction when the backend supports it.
func RollbackMigration(ctx context.Context, conn DB, version string) error {
	var migration *MigrationInfo
	for i := range registeredMigrations {
		if registeredMigrations[i].Version == version {
			migration = &registeredMigrations[i]
			break
		}
	}
	if migration == nil || migration.Down == nil {
		return fmt.Errorf("migration %s is not registered with a rollback", version)
	}

	model := models.GetMigrationModel()
	var existing models.Migration
	if err := conn.FindOne(ctx, model, bson.M{model.VersionKey: version}).Decode(&existing); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("migration %s is not applied", version)
		}
		return fmt.Errorf("failed to check migration %s: %w", version, err)
	}

	log.Infof("Rolling back migration %s: %s", version, migration.Name)
	rollback := func(txCtx context.Context) error {
		if err := migration.Down(txCtx, conn); err != nil {
			return fmt.Errorf("failed to roll back migration %s: %w", version, err)
		}
		if _, err := conn.DeleteOne(txCtx, model, bson.M{model.VersionKey: version}); err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", version, err)
		}
		return nil
	}
	if tx, ok := conn.(TransactionalDB); ok {
		if err := tx.WithTransaction(ctx, rollback); err != nil {
			return err
		}
	} else {
		log.Warnf("Database does not support transactions; rollback of %s and its record are not atomic", version)
		if err := rollback(ctx); err != nil {
			return err
		}
	}
	log.Infof("Successfully rolled back migration %s", version)
	return nil
}

// IsMigrationApplied checks if a specific migration has been applied to the database.
func IsMigrationApplied(ctx context.Context, db DB, version string) (bool, error) {
	migrationModel := models.GetMigrationModel()
	filter := bson.M{migrationModel.VersionKey: version}

	count, err := db.CountDocuments(ctx, migrationModel, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check migration %s: %w", version, err)
	}

	return count > 0, nil
}

// GetMigrations returns a copy of all registered migrations.
func GetMigrations() []MigrationInfo {
	return registeredMigrations
}

// ResetMigrations clears all registered migrations (primarily for tests).
func ResetMigrations() {
	registeredMigrations = []MigrationInfo{}
}
