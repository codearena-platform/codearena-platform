package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type ClientRepository interface {
	Create(ctx context.Context, client *entity.Client) error
	Update(ctx context.Context, client *entity.Client) error
	GetById(ctx context.Context, id string) (*entity.Client, error)
	GetAll(ctx context.Context, query sdk.ClientQueryParams) ([]entity.Client, error)
}

type clientRepository struct {
	db db.DB
}

func NewClientRepository(db db.DB) ClientRepository {
	return &clientRepository{
		db: db,
	}
}

func (r *clientRepository) Create(ctx context.Context, client *entity.Client) error {
	// Hash secret if provided
	if client.Secret != "" {
		hashed, err := hashSecret(client.Secret)
		if err != nil {
			return fmt.Errorf("failed to hash client secret: %w", err)
		}
		client.Secret = hashed
	}

	m := fromEntityClientToModel(*client)
	md := models.GetClientModel()
	_, err := r.db.InsertOne(ctx, md, m)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

func (r *clientRepository) Update(ctx context.Context, client *entity.Client) error {
	m := fromEntityClientToModel(*client)
	md := models.GetClientModel()
	_, err := r.db.UpdateOne(
		ctx,
		md,
		bson.D{{Key: md.IdKey, Value: client.Id}},
		bson.D{{Key: "$set", Value: m}},
	)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}
	return nil
}

func (r *clientRepository) GetById(ctx context.Context, id string) (*entity.Client, error) {
	md := models.GetClientModel()
	var m models.Client
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.IdKey, Value: id}}).Decode(&m)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrClientNotFound
		}
		return nil, fmt.Errorf("failed to find client: %w", err)
	}
	return fromModelToEntityClient(&m), nil
}

func (r *clientRepository) GetAll(ctx context.Context, query sdk.ClientQueryParams) ([]entity.Client, error) {
	md := models.GetClientModel()
	filter := bson.D{}

	if len(query.ProjectIds) > 0 {
		filter = append(filter, bson.E{Key: md.ProjectIdKey, Value: bson.D{{Key: "$in", Value: query.ProjectIds}}})
	}

	if query.IAMClient {
		filter = append(filter, bson.E{Key: md.IAMClientKey, Value: query.IAMClient})
	}

	opts := &options.FindOptions{}
	if query.SortByUpdatedAt {
		opts.SetSort(bson.D{{Key: md.UpdatedAtKey, Value: -1}})
	}

	cursor, err := r.db.Find(ctx, md, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find clients: %w", err)
	}
	defer cursor.Close(ctx)

	var modelsList []models.Client
	if err := cursor.All(ctx, &modelsList); err != nil {
		return nil, fmt.Errorf("failed to decode clients: %w", err)
	}

	clients := make([]entity.Client, len(modelsList))
	for i := range modelsList {
		clients[i] = *fromModelToEntityClient(&modelsList[i])
	}

	return clients, nil
}

// Helpers

func hashSecret(secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
	}
	hashedSecret := sha256.Sum256([]byte(secret))
	return base64.StdEncoding.EncodeToString(hashedSecret[:]), nil
}

func fromEntityClientToModel(e entity.Client) models.Client {
	return models.Client{
		Id:                    e.Id,
		Name:                  e.Name,
		Description:           e.Description,
		Secret:                e.Secret,
		Tags:                  e.Tags,
		RedirectURLs:          e.RedirectURLs,
		DefaultAuthProviderId: e.DefaultAuthProviderId,
		IAMClient:             e.IAMClient,
		ProjectId:             e.ProjectId,
		ServiceAccountEmail:   e.ServiceAccountEmail,
		Scopes:                e.Scopes,
		Enabled:               e.Enabled,
		LinkedUserId:          e.LinkedUserId,
		CreatedAt:             e.CreatedAt,
		CreatedBy:             e.CreatedBy,
		UpdatedAt:             e.UpdatedAt,
		UpdatedBy:             e.UpdatedBy,
	}
}

func fromModelToEntityClient(m *models.Client) *entity.Client {
	if m == nil {
		return nil
	}
	return &entity.Client{
		Id:                    m.Id,
		Name:                  m.Name,
		Description:           m.Description,
		Secret:                m.Secret,
		Tags:                  m.Tags,
		RedirectURLs:          m.RedirectURLs,
		DefaultAuthProviderId: m.DefaultAuthProviderId,
		IAMClient:             m.IAMClient,
		ProjectId:             m.ProjectId,
		ServiceAccountEmail:   m.ServiceAccountEmail,
		Scopes:                m.Scopes,
		Enabled:               m.Enabled,
		LinkedUserId:          m.LinkedUserId,
		CreatedAt:             m.CreatedAt,
		CreatedBy:             m.CreatedBy,
		UpdatedAt:             m.UpdatedAt,
		UpdatedBy:             m.UpdatedBy,
	}
}

func GenerateRandomSecret(length int) (string, error) {
	byteLength := (length*6 + 7) / 8
	bytes := make([]byte, byteLength)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}
