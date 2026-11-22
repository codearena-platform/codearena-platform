package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func createTestMigration(version, name string, shouldFail bool) db.MigrationInfo {
	return db.MigrationInfo{
		Version:     version,
		Name:        name,
		Description: "Test migration " + version,
		Up: func(ctx context.Context, db db.DB) error {
			if shouldFail {
				return errors.New("migration failed")
			}
			return nil
		},
		Down: func(ctx context.Context, db db.DB) error {
			if shouldFail {
				return errors.New("rollback failed")
			}
			return nil
		},
	}
}

func TestRegisterMigration(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	migration := createTestMigration("001", "test_migration", false)
	db.RegisterMigration(migration)

	existingMigrations := db.GetMigrations()

	assert.Len(t, existingMigrations, 1)
	assert.Equal(t, "001", existingMigrations[0].Version)
	assert.Equal(t, "test_migration", existingMigrations[0].Name)
}

func TestCheckAndRunMigrations_NewMigration(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Setup test
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	migrationModel := models.GetMigrationModel()

	// Register test migration
	db.RegisterMigration(createTestMigration("001", "test_migration", false))

	// Mock FindOne to return ErrNoDocuments (migration not found)
	mockResult := mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil)
	mockDB.On("FindOne", ctx, migrationModel, bson.M{migrationModel.VersionKey: "001"}, mock.Anything).Return(mockResult)

	// Mock InsertOne for recording migration
	insertResult := &mongo.InsertOneResult{InsertedID: "001"}
	mockDB.On("InsertOne", ctx, migrationModel, mock.AnythingOfType("models.Migration"), mock.Anything).Return(insertResult, nil)

	// Execute
	err := db.CheckAndRunMigrations(ctx, mockDB)

	// Assert
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestCheckAndRunMigrations_ExistingMigration(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Setup test
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	migrationModel := models.GetMigrationModel()

	// Register test migration
	db.RegisterMigration(createTestMigration("001", "test_migration", false))

	// Mock FindOne to return existing migration (no error)
	mockResult := mongo.NewSingleResultFromDocument(bson.D{}, nil, nil)
	mockDB.On("FindOne", ctx, migrationModel, bson.M{migrationModel.VersionKey: "001"}, mock.Anything).Return(mockResult)

	// Execute
	err := db.CheckAndRunMigrations(ctx, mockDB)

	// Assert
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	// InsertOne should not be called since migration already exists
	mockDB.AssertNotCalled(t, "InsertOne")
}

func TestCheckAndRunMigrations_MigrationUpFails(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Setup test
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	migrationModel := models.GetMigrationModel()

	// Register test migration that fails
	db.RegisterMigration(createTestMigration("001", "test_migration", true))

	// Mock FindOne to return ErrNoDocuments (migration not found)
	mockResult := mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil)
	mockDB.On("FindOne", ctx, migrationModel, bson.M{migrationModel.VersionKey: "001"}, mock.Anything).Return(mockResult)

	// Execute
	err := db.CheckAndRunMigrations(ctx, mockDB)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to apply migration 001")
	mockDB.AssertExpectations(t)
	// InsertOne should not be called since migration failed
	mockDB.AssertNotCalled(t, "InsertOne")
}

func TestCheckAndRunMigrations_FindOneFails(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Setup test
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	migrationModel := models.GetMigrationModel()

	// Register test migration
	db.RegisterMigration(createTestMigration("001", "test_migration", false))

	// Mock FindOne to return a different error
	mockResult := mongo.NewSingleResultFromDocument(bson.D{}, errors.New("database error"), nil)
	mockDB.On("FindOne", ctx, migrationModel, bson.M{migrationModel.VersionKey: "001"}, mock.Anything).Return(mockResult)

	// Execute
	err := db.CheckAndRunMigrations(ctx, mockDB)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check migration 001: database error")
	mockDB.AssertExpectations(t)
}

func TestCheckAndRunMigrations_InsertOneFails(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Setup test
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	migrationModel := models.GetMigrationModel()

	// Register test migration
	db.RegisterMigration(createTestMigration("001", "test_migration", false))

	// Mock FindOne to return ErrNoDocuments (migration not found)
	mockResult := mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil)
	mockDB.On("FindOne", ctx, migrationModel, bson.M{migrationModel.VersionKey: "001"}, mock.Anything).Return(mockResult)

	// Mock InsertOne to fail
	mockDB.On("InsertOne", ctx, migrationModel, mock.AnythingOfType("models.Migration"), mock.Anything).Return((*mongo.InsertOneResult)(nil), errors.New("insert failed"))

	// Execute
	err := db.CheckAndRunMigrations(ctx, mockDB)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record migration 001: insert failed")
	mockDB.AssertExpectations(t)
}

func TestCheckAndRunMigrations_MultipleMigrations(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Setup test
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	migrationModel := models.GetMigrationModel()

	// Register multiple test migrations
	db.RegisterMigration(createTestMigration("001", "first_migration", false))
	db.RegisterMigration(createTestMigration("002", "second_migration", false))

	// Mock FindOne for first migration (not found)
	mockResult1 := mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil)
	mockDB.On("FindOne", ctx, migrationModel, bson.M{migrationModel.VersionKey: "001"}, mock.Anything).Return(mockResult1)

	// Mock FindOne for second migration (already exists)
	mockResult2 := mongo.NewSingleResultFromDocument(bson.D{}, nil, nil)
	mockDB.On("FindOne", ctx, migrationModel, bson.M{migrationModel.VersionKey: "002"}, mock.Anything).Return(mockResult2)

	// Mock InsertOne for first migration only
	insertResult := &mongo.InsertOneResult{InsertedID: "001"}
	mockDB.On("InsertOne", ctx, migrationModel, mock.AnythingOfType("models.Migration"), mock.Anything).Return(insertResult, nil)

	// Execute
	err := db.CheckAndRunMigrations(ctx, mockDB)

	// Assert
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestIsMigrationApplied_Exists(t *testing.T) {
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	version := "001"
	migrationModel := models.GetMigrationModel()

	// Mock CountDocuments to return 1 (migration exists)
	mockDB.On("CountDocuments", ctx, migrationModel, bson.M{migrationModel.VersionKey: version}, mock.Anything).Return(int64(1), nil)

	// Execute
	applied, err := db.IsMigrationApplied(ctx, mockDB, version)

	// Assert
	assert.NoError(t, err)
	assert.True(t, applied)
	mockDB.AssertExpectations(t)
}

func TestIsMigrationApplied_NotExists(t *testing.T) {
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	version := "001"
	migrationModel := models.GetMigrationModel()

	// Mock CountDocuments to return 0 (migration does not exist)
	mockDB.On("CountDocuments", ctx, migrationModel, bson.M{migrationModel.VersionKey: version}, mock.Anything).Return(int64(0), nil)

	// Execute
	applied, err := db.IsMigrationApplied(ctx, mockDB, version)

	// Assert
	assert.NoError(t, err)
	assert.False(t, applied)
	mockDB.AssertExpectations(t)
}

func TestIsMigrationApplied_Error(t *testing.T) {
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	version := "001"
	migrationModel := models.GetMigrationModel()

	// Mock CountDocuments to return error
	mockDB.On("CountDocuments", ctx, migrationModel, bson.M{migrationModel.VersionKey: version}, mock.Anything).Return(int64(0), errors.New("database error"))

	// Execute
	applied, err := db.IsMigrationApplied(ctx, mockDB, version)

	// Assert
	assert.Error(t, err)
	assert.False(t, applied)
	assert.Contains(t, err.Error(), "failed to check migration 001")
	mockDB.AssertExpectations(t)
}

func TestCheckAndRunMigrations_EmptyMigrations(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Setup test with no migrations
	mockDB := test.SetupMockDB()
	ctx := context.Background()

	// Execute
	err := db.CheckAndRunMigrations(ctx, mockDB)

	// Assert
	assert.NoError(t, err)
	// No database calls should be made
	mockDB.AssertNotCalled(t, "FindOne")
	mockDB.AssertNotCalled(t, "InsertOne")
}

func TestMigrationInfo_Structure(t *testing.T) {
	migration := db.MigrationInfo{
		Version:     "001",
		Name:        "test_migration",
		Description: "A test migration",
		Up: func(ctx context.Context, db db.DB) error {
			return nil
		},
		Down: func(ctx context.Context, db db.DB) error {
			return nil
		},
	}

	assert.Equal(t, "001", migration.Version)
	assert.Equal(t, "test_migration", migration.Name)
	assert.Equal(t, "A test migration", migration.Description)
	assert.NotNil(t, migration.Up)
	assert.NotNil(t, migration.Down)

	// Test that functions can be called
	err := migration.Up(context.Background(), nil)
	assert.NoError(t, err)

	err = migration.Down(context.Background(), nil)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkRegisterMigration(b *testing.B) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	migration := createTestMigration("001", "benchmark_migration", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ResetMigrations()
		db.RegisterMigration(migration)
	}
}

func BenchmarkIsMigrationApplied(b *testing.B) {
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	version := "001"
	migrationModel := models.GetMigrationModel()

	mockDB.On("CountDocuments", ctx, migrationModel, mock.Anything, mock.Anything).Return(int64(1), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok, err := db.IsMigrationApplied(ctx, mockDB, version)
		assert.NoError(b, err, "unexpected error")
		assert.True(b, ok, "expected migration to be applied")
	}
}

func TestGetMigrations(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Reset migrations for test
	db.ResetMigrations()

	// Register some test migrations
	migration1 := createTestMigration("001", "first", false)
	migration2 := createTestMigration("002", "second", false)
	db.RegisterMigration(migration1)
	db.RegisterMigration(migration2)

	// Get migrations
	migrations := db.GetMigrations()

	// Assert
	assert.Len(t, migrations, 2)
	assert.Equal(t, "001", migrations[0].Version)
	assert.Equal(t, "002", migrations[1].Version)
}

type fakeTransactionalDB struct {
	*test.MockDB
	transactions int
	commits      int
	aborts       int
}

func (f *fakeTransactionalDB) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	f.transactions++
	if err := fn(ctx); err != nil {
		f.aborts++
		return err
	}
	f.commits++
	return nil
}

func withIsolatedMigrations(t *testing.T) {
	t.Helper()
	original := db.GetMigrations()
	t.Cleanup(func() {
		db.ResetMigrations()
		for _, migration := range original {
			db.RegisterMigration(migration)
		}
	})
	db.ResetMigrations()
}

func TestCheckAndRunMigrations_UsesTransaction(t *testing.T) {
	withIsolatedMigrations(t)
	mockDB := &fakeTransactionalDB{MockDB: test.SetupMockDB()}
	ctx := context.Background()
	model := models.GetMigrationModel()

	db.RegisterMigration(createTestMigration("001", "transactional", false))
	mockDB.On("FindOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
		Return(mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil))
	mockDB.On("InsertOne", ctx, model, mock.AnythingOfType("models.Migration"), mock.Anything).
		Return(&mongo.InsertOneResult{InsertedID: "001"}, nil)

	require.NoError(t, db.CheckAndRunMigrations(ctx, mockDB))
	assert.Equal(t, 1, mockDB.transactions)
	assert.Equal(t, 1, mockDB.commits)
	assert.Equal(t, 0, mockDB.aborts)
}

func TestCheckAndRunMigrations_TransactionRollsBackOnFailure(t *testing.T) {
	withIsolatedMigrations(t)
	mockDB := &fakeTransactionalDB{MockDB: test.SetupMockDB()}
	ctx := context.Background()
	model := models.GetMigrationModel()

	db.RegisterMigration(createTestMigration("001", "failing", true))
	mockDB.On("FindOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
		Return(mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil))

	err := db.CheckAndRunMigrations(ctx, mockDB)
	assert.ErrorContains(t, err, "failed to apply migration 001")
	assert.Equal(t, 1, mockDB.transactions)
	assert.Equal(t, 1, mockDB.aborts)
	mockDB.AssertNotCalled(t, "InsertOne")
}

func TestCheckAndRunMigrations_RecordsChecksumAndDetectsDrift(t *testing.T) {
	withIsolatedMigrations(t)
	ctx := context.Background()
	model := models.GetMigrationModel()
	migration := createTestMigration("001", "checksummed", false)
	db.RegisterMigration(migration)

	t.Run("records a checksum", func(t *testing.T) {
		mockDB := test.SetupMockDB()
		mockDB.On("FindOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
			Return(mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil))
		mockDB.On("InsertOne", ctx, model, mock.MatchedBy(func(record models.Migration) bool {
			return len(record.Checksum) == 64
		}), mock.Anything).Return(&mongo.InsertOneResult{InsertedID: "001"}, nil)

		require.NoError(t, db.CheckAndRunMigrations(ctx, mockDB))
		mockDB.AssertExpectations(t)
	})

	t.Run("rejects a changed migration", func(t *testing.T) {
		mockDB := test.SetupMockDB()
		mockDB.On("FindOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
			Return(mongo.NewSingleResultFromDocument(models.Migration{Checksum: "tampered"}, nil, nil))

		err := db.CheckAndRunMigrations(ctx, mockDB)
		assert.ErrorContains(t, err, "checksum mismatch")
		mockDB.AssertNotCalled(t, "InsertOne")
	})

	t.Run("accepts legacy records without a checksum", func(t *testing.T) {
		mockDB := test.SetupMockDB()
		mockDB.On("FindOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
			Return(mongo.NewSingleResultFromDocument(bson.D{}, nil, nil))

		assert.NoError(t, db.CheckAndRunMigrations(ctx, mockDB))
	})
}

func TestCheckAndRunMigrations_DryRun(t *testing.T) {
	withIsolatedMigrations(t)
	mockDB := test.SetupMockDB()
	ctx := context.Background()
	model := models.GetMigrationModel()

	applied := false
	db.RegisterMigration(db.MigrationInfo{
		Version: "001",
		Name:    "dry_run",
		Up: func(context.Context, db.DB) error {
			applied = true
			return nil
		},
		Down: func(context.Context, db.DB) error { return nil },
	})
	mockDB.On("FindOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
		Return(mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil))

	require.NoError(t, db.CheckAndRunMigrations(ctx, mockDB, db.WithDryRun(true)))
	assert.False(t, applied)
	mockDB.AssertNotCalled(t, "InsertOne")
	mockDB.AssertNotCalled(t, "UpdateOne")
}

func TestRollbackMigration(t *testing.T) {
	withIsolatedMigrations(t)
	ctx := context.Background()
	model := models.GetMigrationModel()

	mockDB := &fakeTransactionalDB{MockDB: test.SetupMockDB()}
	rolledBack := false
	db.RegisterMigration(db.MigrationInfo{
		Version: "001",
		Name:    "rollback",
		Up:      func(context.Context, db.DB) error { return nil },
		Down: func(context.Context, db.DB) error {
			rolledBack = true
			return nil
		},
	})
	mockDB.On("FindOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
		Return(mongo.NewSingleResultFromDocument(bson.D{}, nil, nil))
	mockDB.On("DeleteOne", ctx, model, bson.M{model.VersionKey: "001"}, mock.Anything).
		Return(&mongo.DeleteResult{DeletedCount: 1}, nil)

	require.NoError(t, db.RollbackMigration(ctx, mockDB, "001"))
	assert.True(t, rolledBack)
	assert.Equal(t, 1, mockDB.commits)

	err := db.RollbackMigration(ctx, mockDB, "missing")
	assert.ErrorContains(t, err, "not registered")
}

func TestResetMigrations(t *testing.T) {
	// Save original state
	originalMigrations := db.GetMigrations()
	defer func() {
		db.ResetMigrations()
		for _, migration := range originalMigrations {
			db.RegisterMigration(migration)
		}
	}()

	// Register a migration
	db.RegisterMigration(createTestMigration("001", "test", false))
	assert.Len(t, db.GetMigrations(), len(originalMigrations)+1)

	// Reset
	db.ResetMigrations()
	assert.Len(t, db.GetMigrations(), 0)
}
