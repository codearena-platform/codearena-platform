package mongo

import (
	"context"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Match document field names. These must match the bson tags on entity.Match;
// match_fields_test.go enforces that invariant.
const (
	matchFieldID           = "id"
	matchFieldTournamentID = "tournament_id"
	matchFieldStatus       = "status"
	matchFieldParticipants = "participants"
	matchFieldWinnerID     = "winner_id"
	matchFieldEndedAt      = "ended_at"
)

type MongoMatchRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoMatchRepository(client *mongo.Client, dbName string) *MongoMatchRepository {
	return &MongoMatchRepository{
		client:     client,
		collection: client.Database(dbName).Collection("matches"),
	}
}

func matchIDFilter(id string) bson.M {
	return bson.M{matchFieldID: id}
}

func finishedMatchesFilter() bson.M {
	return bson.M{matchFieldStatus: bson.M{"$in": []entity.MatchStatus{entity.MatchFinished, entity.MatchCompleted}}}
}

func finishedMatchesSort() bson.M {
	return bson.M{matchFieldEndedAt: -1}
}

func botMatchesFilter(botID string) bson.M {
	return bson.M{matchFieldParticipants: botID}
}

func tournamentMatchesFilter(tournamentID string) bson.M {
	return bson.M{matchFieldTournamentID: tournamentID}
}

func aggregateStatsPipeline() mongo.Pipeline {
	return mongo.Pipeline{
		{{Key: "$match", Value: bson.M{matchFieldStatus: entity.MatchCompleted}}},
		{{Key: "$group", Value: bson.M{
			"_id":           nil,
			"total_matches": bson.M{"$sum": 1},
			"winners":       bson.M{"$push": "$" + matchFieldWinnerID},
		}}},
	}
}

func (r *MongoMatchRepository) Save(match *entity.Match) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	filter := matchIDFilter(match.ID)
	update := bson.M{"$set": match}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoMatchRepository) GetByID(id string) (*entity.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var match entity.Match
	err := r.collection.FindOne(ctx, matchIDFilter(id)).Decode(&match)
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MongoMatchRepository) ListActive() ([]*entity.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{matchFieldStatus: entity.MatchRunning}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matches []*entity.Match
	if err = cursor.All(ctx, &matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *MongoMatchRepository) ListFinished(limit int) ([]*entity.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetLimit(int64(limit)).SetSort(finishedMatchesSort())

	cursor, err := r.collection.Find(ctx, finishedMatchesFilter(), opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matches []*entity.Match
	if err = cursor.All(ctx, &matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *MongoMatchRepository) GetByBotID(botID string) ([]*entity.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := botMatchesFilter(botID)
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matches []*entity.Match
	if err = cursor.All(ctx, &matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *MongoMatchRepository) GetByTournamentID(tournamentID string) ([]*entity.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := tournamentMatchesFilter(tournamentID)
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var matches []*entity.Match
	if err = cursor.All(ctx, &matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *MongoMatchRepository) GetAggregateStats(query map[string]any) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Aggregate(ctx, aggregateStatsPipeline())
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return map[string]any{"total_matches": 0}, nil
	}

	return results[0], nil
}
