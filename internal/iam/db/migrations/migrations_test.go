package migrations

import (
	"context"
	"fmt"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/policy/system"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// seededDB is a tiny in-memory DB used to exercise the migration against real
// BSON documents rather than only asserting on mock calls.
type seededDB struct {
	db.DB
	users []bson.M
}

func (s *seededDB) Find(_ context.Context, col db.DbCollection, _ interface{}, _ ...*options.FindOptions) (*mongo.Cursor, error) {
	if col.Name() != "users" {
		return nil, fmt.Errorf("unexpected collection %s", col.Name())
	}
	docs := make([]interface{}, 0, len(s.users))
	for _, user := range s.users {
		docs = append(docs, user)
	}
	return mongo.NewCursorFromDocuments(docs, nil, nil)
}

func (s *seededDB) UpdateOne(_ context.Context, col db.DbCollection, filter interface{}, update interface{}, _ ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if col.Name() != "users" {
		return nil, fmt.Errorf("unexpected collection %s", col.Name())
	}
	wanted := filter.(bson.M)["id"]
	changes := update.(bson.M)
	for _, document := range s.users {
		if document["id"] != wanted {
			continue
		}
		if set, ok := changes["$set"].(bson.M); ok {
			for key, value := range set {
				setField(document, key, value)
			}
		}
		if unset, ok := changes["$unset"].(bson.M); ok {
			for key := range unset {
				setField(document, key, nil)
			}
		}
		return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
	}
	return &mongo.UpdateResult{}, nil
}

func setField(document bson.M, key string, value interface{}) {
	parts := splitKey(key)
	if len(parts) == 1 {
		if value == nil {
			delete(document, parts[0])
		} else {
			document[parts[0]] = value
		}
		return
	}
	switch parent := document[parts[0]].(type) {
	case bson.M:
		if value == nil {
			delete(parent, parts[1])
		} else {
			parent[parts[1]] = value
		}
	case map[string]models.UserPolicy:
		if value == nil {
			delete(parent, parts[1])
		} else if policy, ok := value.(models.UserPolicy); ok {
			parent[parts[1]] = policy
		}
	default:
		nested := bson.M{}
		if value != nil {
			nested[parts[1]] = value
		}
		document[parts[0]] = nested
	}
}

func splitKey(key string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	return append(parts, key[start:])
}

func seededUser(id string, policies interface{}) bson.M {
	document := bson.M{"id": id, "name": id}
	if policies != nil {
		document["policies"] = policies
	}
	return document
}

func userPolicies(t *testing.T, document bson.M) map[string]models.UserPolicy {
	t.Helper()
	raw, err := bson.Marshal(document["policies"])
	require.NoError(t, err)
	var policies map[string]models.UserPolicy
	require.NoError(t, bson.Unmarshal(raw, &policies))
	return policies
}

func legacyPolicies(t *testing.T, document bson.M) map[string]string {
	t.Helper()
	raw, err := bson.Marshal(document["policies"])
	require.NoError(t, err)
	var policies map[string]string
	require.NoError(t, bson.Unmarshal(raw, &policies))
	return policies
}

func TestMigration_UpdateUserPoliciesUp(t *testing.T) {
	ctx := context.Background()
	addedID := system.NewAccessToCreatedResource(nil).ID()

	t.Run("preserves legacy policies and adds the new policy", func(t *testing.T) {
		conn := &seededDB{users: []bson.M{
			seededUser("user-1", map[string]string{"legacy-policy": "AllowRead"}),
		}}

		require.NoError(t, updateUserPoliciesUp(ctx, conn))

		policies := userPolicies(t, conn.users[0])
		assert.Equal(t, "AllowRead", policies["legacy-policy"].Name)
		assert.Contains(t, policies, addedID)
		assert.Contains(t, conn.users[0], migrationBackupField)
	})

	t.Run("preserves structured policies and is idempotent", func(t *testing.T) {
		custom := models.UserPolicy{Name: "CustomPolicy"}
		conn := &seededDB{users: []bson.M{
			seededUser("user-2", map[string]models.UserPolicy{"custom": custom}),
		}}

		require.NoError(t, updateUserPoliciesUp(ctx, conn))
		first := userPolicies(t, conn.users[0])
		assert.Equal(t, "CustomPolicy", first["custom"].Name)
		assert.Contains(t, first, addedID)

		// A second run must not overwrite newer changes made after migration.
		first["added-later"] = models.UserPolicy{Name: "AddedLater"}
		conn.users[0]["policies"] = first
		require.NoError(t, updateUserPoliciesUp(ctx, conn))
		assert.Contains(t, userPolicies(t, conn.users[0]), "added-later")
	})

	t.Run("handles users without policies", func(t *testing.T) {
		conn := &seededDB{users: []bson.M{seededUser("user-3", nil)}}

		require.NoError(t, updateUserPoliciesUp(ctx, conn))
		assert.Contains(t, userPolicies(t, conn.users[0]), addedID)
		backup := conn.users[0][migrationBackupField].(bson.M)
		assert.Equal(t, false, backup["existed"])
	})
}

func TestMigration_UpdateUserPoliciesDown(t *testing.T) {
	ctx := context.Background()
	addedID := system.NewAccessToCreatedResource(nil).ID()

	t.Run("restores the exact legacy policies", func(t *testing.T) {
		conn := &seededDB{users: []bson.M{
			seededUser("user-1", map[string]string{"legacy-policy": "AllowRead"}),
		}}
		require.NoError(t, updateUserPoliciesUp(ctx, conn))

		require.NoError(t, updateUserPoliciesDown(ctx, conn))

		assert.Equal(t, map[string]string{"legacy-policy": "AllowRead"}, legacyPolicies(t, conn.users[0]))
		assert.NotContains(t, conn.users[0], migrationBackupField)
	})

	t.Run("removes policies when the user had none", func(t *testing.T) {
		conn := &seededDB{users: []bson.M{seededUser("user-3", nil)}}
		require.NoError(t, updateUserPoliciesUp(ctx, conn))
		require.NoError(t, updateUserPoliciesDown(ctx, conn))

		assert.NotContains(t, conn.users[0], "policies")
		assert.NotContains(t, conn.users[0], migrationBackupField)
	})

	t.Run("removes only the added policy when no backup exists", func(t *testing.T) {
		conn := &seededDB{users: []bson.M{
			seededUser("user-4", map[string]models.UserPolicy{
				addedID: {Name: system.NewAccessToCreatedResource(nil).Name()},
				"keep":  {Name: "KeepMe"},
			}),
		}}

		require.NoError(t, updateUserPoliciesDown(ctx, conn))

		policies := userPolicies(t, conn.users[0])
		assert.NotContains(t, policies, addedID)
		assert.Equal(t, "KeepMe", policies["keep"].Name)
	})
}

func TestMigration_Registration(t *testing.T) {
	migrations := db.GetMigrations()
	var found db.MigrationInfo
	for _, migration := range migrations {
		if migration.Version == "001" {
			found = migration
		}
	}
	assert.Equal(t, "001", found.Version)
	assert.Equal(t, "update_user_policies", found.Name)
	assert.NotNil(t, found.Up)
	assert.NotNil(t, found.Down)
}
