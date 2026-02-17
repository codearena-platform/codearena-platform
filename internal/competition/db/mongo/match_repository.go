package mongo

import (
	"context"
	"time"

	"github.com/codearena-platform/codearena-platform/internal/competition/services/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *MongoMatchRepository) Save(match *entity.Match) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"id": match.ID}
	update := bson.M{"$set": match}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoMatchRepository) GetByID(id string) (*entity.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var match entity.Match
	filter := bson.M{"id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&match)
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MongoMatchRepository) ListActive() ([]*entity.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"status": entity.MatchRunning}
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

	filter := bson.M{"status": bson.M{"$in": []entity.MatchStatus{entity.MatchFinished, entity.MatchCompleted}}}
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.M{"ended_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
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

	filter := bson.M{"participants": botID}
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

	filter := bson.M{"tournament_id": tournamentID}
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

	// Initial basic aggregation: Count total matches and group by winner
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": entity.MatchCompleted}}},
		{{Key: "$group", Value: bson.M{
			"_id":           nil,
			"total_matches": bson.M{"$sum": 1},
			"winners":       bson.M{"$push": "$winner_id"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
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
