package mongo

import (
	"reflect"
	"testing"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMatchBSONTagsMatchRepositoryFields(t *testing.T) {
	expected := map[string]string{
		"ID":              matchFieldID,
		"TournamentID":    matchFieldTournamentID,
		"Status":          matchFieldStatus,
		"Participants":    matchFieldParticipants,
		"WinnerID":        matchFieldWinnerID,
		"EndedAt":         matchFieldEndedAt,
		"DivisionID":      "division_id",
		"ChallengeID":     "challenge_id",
		"StartedAt":       "started_at",
		"FinishedAt":      "finished_at",
		"EngineURL":       "engine_url",
		"ReplayAvailable": "replay_available",
		"Metadata":        "metadata",
	}

	matchType := reflect.TypeOf(entity.Match{})
	for field, tag := range expected {
		structField, ok := matchType.FieldByName(field)
		require.True(t, ok, "entity.Match.%s is missing", field)
		assert.Equal(t, tag, structField.Tag.Get("bson"), "entity.Match.%s", field)
	}
}

func TestMatchFiltersUseStoredFieldNames(t *testing.T) {
	assert.Equal(t, bson.M{matchFieldID: "match-1"}, matchIDFilter("match-1"))
	assert.Equal(t, bson.M{matchFieldParticipants: "bot-1"}, botMatchesFilter("bot-1"))
	assert.Equal(t, bson.M{matchFieldTournamentID: "tournament-1"}, tournamentMatchesFilter("tournament-1"))
	assert.Equal(t, bson.M{matchFieldEndedAt: -1}, finishedMatchesSort())

	statusFilter := finishedMatchesFilter()[matchFieldStatus]
	assert.Equal(t, bson.M{"$in": []entity.MatchStatus{entity.MatchFinished, entity.MatchCompleted}}, statusFilter)
}

func TestAggregateStatsUsesWinnerField(t *testing.T) {
	pipeline := aggregateStatsPipeline()
	require.Len(t, pipeline, 2)

	matchStage := pipeline[0]
	require.Len(t, matchStage, 1)
	assert.Equal(t, "$match", matchStage[0].Key)
	assert.Equal(t, entity.MatchCompleted, matchStage[0].Value.(bson.M)[matchFieldStatus])

	groupStage := pipeline[1]
	require.Len(t, groupStage, 1)
	assert.Equal(t, "$group", groupStage[0].Key)
	assert.Equal(t, "$"+matchFieldWinnerID, groupStage[0].Value.(bson.M)["winners"].(bson.M)["$push"])
}
