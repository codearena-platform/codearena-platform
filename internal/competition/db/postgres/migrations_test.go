package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMigrationStore struct {
	applied map[int]bool
	calls   []int
	failOn  int
}

func (s *fakeMigrationStore) AppliedVersions(context.Context) (map[int]bool, error) {
	return s.applied, nil
}

func (s *fakeMigrationStore) Apply(_ context.Context, m migration) error {
	s.calls = append(s.calls, m.Version)
	if m.Version == s.failOn {
		return errors.New("apply failed")
	}
	s.applied[m.Version] = true
	return nil
}

func TestRunMigrationsAppliesAllVersionsInOrder(t *testing.T) {
	store := &fakeMigrationStore{applied: map[int]bool{}}

	require.NoError(t, runMigrations(context.Background(), store))
	assert.Equal(t, []int{1, 2}, store.calls)
	assert.Equal(t, map[int]bool{1: true, 2: true}, store.applied)
}

func TestRunMigrationsSkipsAppliedVersions(t *testing.T) {
	store := &fakeMigrationStore{applied: map[int]bool{1: true}}

	require.NoError(t, runMigrations(context.Background(), store))
	assert.Equal(t, []int{2}, store.calls)
}

func TestRunMigrationsStopsOnFailure(t *testing.T) {
	store := &fakeMigrationStore{applied: map[int]bool{}, failOn: 1}

	err := runMigrations(context.Background(), store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "apply failed")
	assert.Equal(t, []int{1}, store.calls)
}

func TestMigrationsCreateEveryTableUsedByRepositories(t *testing.T) {
	var statements []string
	for _, m := range migrations {
		statements = append(statements, m.Statements...)
	}
	schema := strings.ToLower(strings.Join(statements, "\n"))

	for _, table := range []string{
		"bots", "scores", "tournaments", "brackets", "leagues", "divisions",
		"standings", "hubs", "hub_memberships",
	} {
		assert.Contains(t, schema, "create table if not exists "+table, "table %s", table)
	}

	for _, column := range []string{
		"hub_id", "owner_id", "visibility", "joined_at", "role",
		"promotion_slots", "relegation_slots", "dano_causado",
	} {
		assert.Contains(t, schema, column, "column %s", column)
	}

	assert.Contains(t, schema, "alter table tournaments add column if not exists hub_id")
	assert.Len(t, migrations, 2)
	assert.Equal(t, []int{1, 2}, []int{migrations[0].Version, migrations[1].Version})
}
