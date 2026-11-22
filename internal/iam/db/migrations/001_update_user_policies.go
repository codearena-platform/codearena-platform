package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/policy/system"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// migrationBackupField keeps the exact pre-migration policies so the
	// rollback can restore them instead of writing an empty map.
	migrationBackupField = "migration_001_policies_backup"
	migratorUp           = "migration_001"
	migratorDown         = "migration_001_rollback"
)

func init() {
	db.RegisterMigration(db.MigrationInfo{
		Version:     "001",
		Name:        "update_user_policies",
		Description: "Convert legacy user policies to the structured format without dropping entries, and add NewAccessToCreatedResource",
		Up:          updateUserPoliciesUp,
		Down:        updateUserPoliciesDown,
	})
}

// convertPolicies accepts both the legacy map[string]string and the structured
// map[string]models.UserPolicy formats. Every legacy entry is carried over;
// reportAlreadyStructured tells the caller when no conversion was needed.
func convertPolicies(value interface{}) (map[string]models.UserPolicy, bool, error) {
	policies := map[string]models.UserPolicy{}
	if value == nil {
		return policies, false, nil
	}
	raw, err := bson.Marshal(value)
	if err != nil {
		return nil, false, fmt.Errorf("encode user policies: %w", err)
	}

	var structured map[string]models.UserPolicy
	if err := bson.Unmarshal(raw, &structured); err == nil {
		for id, policy := range structured {
			policies[id] = policy
		}
		return policies, true, nil
	}

	var legacy map[string]string
	if err := bson.Unmarshal(raw, &legacy); err != nil {
		return nil, false, fmt.Errorf("unsupported user policies format: %w", err)
	}
	for id, name := range legacy {
		policies[id] = models.UserPolicy{Name: name}
	}
	return policies, false, nil
}

func updateUserPoliciesUp(ctx context.Context, conn db.DB) error {
	userModel := models.GetUserModel()
	accessPolicy := system.NewAccessToCreatedResource(nil)
	addedID := accessPolicy.ID()
	addedPolicy := models.UserPolicy{Name: accessPolicy.Name()}

	log.Info("Starting user policies migration...")

	// Sort by _id and stream the cursor: skip/limit pagination can silently
	// skip documents that were just updated by the migration itself.
	findOpts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := conn.Find(ctx, userModel, bson.M{}, findOpts)
	if err != nil {
		return fmt.Errorf("failed to find users: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Errorf("failed to close cursor: %w", err)
		}
	}()

	processed := 0
	for cursor.Next(ctx) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			return fmt.Errorf("failed to decode user: %w", err)
		}
		userID, _ := document[userModel.IdKey].(string)
		if userID == "" {
			return fmt.Errorf("found user document without %s", userModel.IdKey)
		}
		// Idempotency: a backup means this document was already migrated.
		if _, alreadyMigrated := document[migrationBackupField]; alreadyMigrated {
			continue
		}

		policies, alreadyStructured, err := convertPolicies(document[userModel.PoliciesKey])
		if err != nil {
			return fmt.Errorf("failed to convert policies for user %s: %w", userID, err)
		}
		if _, exists := policies[addedID]; exists && alreadyStructured {
			continue
		}
		if _, exists := policies[addedID]; !exists {
			policies[addedID] = addedPolicy
		}

		backup := bson.M{"existed": document[userModel.PoliciesKey] != nil}
		if document[userModel.PoliciesKey] != nil {
			backup["policies"] = document[userModel.PoliciesKey]
		}
		now := time.Now()
		update := bson.M{"$set": bson.M{
			userModel.PoliciesKey: policies,
			migrationBackupField:  backup,
			"updated_at":          &now,
			"updated_by":          migratorUp,
		}}
		if _, err := conn.UpdateOne(ctx, userModel, bson.M{userModel.IdKey: userID}, update); err != nil {
			return fmt.Errorf("failed to update user %s: %w", userID, err)
		}
		processed++
	}
	if err := cursor.Err(); err != nil {
		return fmt.Errorf("failed while iterating users: %w", err)
	}

	log.Infof("Successfully migrated %d users", processed)
	return nil
}

func updateUserPoliciesDown(ctx context.Context, conn db.DB) error {
	userModel := models.GetUserModel()
	accessPolicy := system.NewAccessToCreatedResource(nil)
	addedID := accessPolicy.ID()
	addedPolicy := models.UserPolicy{Name: accessPolicy.Name()}

	log.Info("Rolling back user policies migration...")

	findOpts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := conn.Find(ctx, userModel, bson.M{}, findOpts)
	if err != nil {
		return fmt.Errorf("failed to find users: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Errorf("failed to close cursor: %w", err)
		}
	}()

	processed := 0
	for cursor.Next(ctx) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			return fmt.Errorf("failed to decode user: %w", err)
		}
		userID, _ := document[userModel.IdKey].(string)
		if userID == "" {
			return fmt.Errorf("found user document without %s", userModel.IdKey)
		}
		filter := bson.M{userModel.IdKey: userID}
		now := time.Now()
		metadata := bson.M{"updated_at": &now, "updated_by": migratorDown}

		backup, ok := document[migrationBackupField].(bson.M)
		if ok {
			set := bson.M{}
			unset := bson.M{migrationBackupField: ""}
			existed, _ := backup["existed"].(bool)
			if existed {
				set[userModel.PoliciesKey] = backup["policies"]
			} else {
				unset[userModel.PoliciesKey] = ""
			}
			update := bson.M{"$set": set, "$unset": unset}
			for k, v := range metadata {
				update["$set"].(bson.M)[k] = v
			}
			if _, err := conn.UpdateOne(ctx, userModel, filter, update); err != nil {
				return fmt.Errorf("failed to roll back user %s: %w", userID, err)
			}
			processed++
			continue
		}

		// Documents migrated by the old version have no backup: the original
		// policies cannot be recovered, but the policy that migration added can
		// be removed without touching anything else.
		policies, _, err := convertPolicies(document[userModel.PoliciesKey])
		if err != nil {
			return fmt.Errorf("failed to inspect policies for user %s: %w", userID, err)
		}
		if policy, exists := policies[addedID]; !exists || policy.Name != addedPolicy.Name {
			continue
		}
		unset := bson.M{"policies." + addedID: ""}
		update := bson.M{"$set": metadata, "$unset": unset}
		if _, err := conn.UpdateOne(ctx, userModel, filter, update); err != nil {
			return fmt.Errorf("failed to roll back user %s: %w", userID, err)
		}
		processed++
	}
	if err := cursor.Err(); err != nil {
		return fmt.Errorf("failed while iterating users: %w", err)
	}

	log.Infof("Successfully rolled back %d users", processed)
	return nil
}
