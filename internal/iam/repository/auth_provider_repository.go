package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/services/encrypt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrAuthProviderNotFound = errors.New("auth provider not found")
)

type AuthProviderRepository interface {
	Create(ctx context.Context, provider *entity.AuthProvider) error
	Update(ctx context.Context, provider *entity.AuthProvider) error
	GetById(ctx context.Context, id string) (*entity.AuthProvider, error)
	GetAll(ctx context.Context, projectIds []string) ([]entity.AuthProvider, error)
}

type authProviderRepository struct {
	db  db.DB
	enc encrypt.Service
}

func NewAuthProviderRepository(db db.DB, enc encrypt.Service) AuthProviderRepository {
	return &authProviderRepository{
		db:  db,
		enc: enc,
	}
}

func (r *authProviderRepository) Create(ctx context.Context, provider *entity.AuthProvider) error {
	m := fromEntityAuthProviderToModel(*provider)

	// Encrypt secrets
	if err := r.encryptSecrets(&m); err != nil {
		return fmt.Errorf("failed to encrypt auth provider secrets: %w", err)
	}

	md := models.GetAuthProviderModel()
	_, err := r.db.InsertOne(ctx, md, m)
	if err != nil {
		return fmt.Errorf("failed to create auth provider: %w", err)
	}

	return nil
}

func (r *authProviderRepository) Update(ctx context.Context, provider *entity.AuthProvider) error {
	m := fromEntityAuthProviderToModel(*provider)

	// Encrypt secrets
	if err := r.encryptSecrets(&m); err != nil {
		return fmt.Errorf("failed to encrypt auth provider secrets: %w", err)
	}

	md := models.GetAuthProviderModel()
	_, err := r.db.UpdateOne(
		ctx,
		md,
		bson.D{{Key: md.IdKey, Value: provider.Id}},
		bson.D{{Key: "$set", Value: m}},
	)
	if err != nil {
		return fmt.Errorf("failed to update auth provider: %w", err)
	}

	return nil
}

func (r *authProviderRepository) GetById(ctx context.Context, id string) (*entity.AuthProvider, error) {
	md := models.GetAuthProviderModel()
	var m models.AuthProvider
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.IdKey, Value: id}}).Decode(&m)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAuthProviderNotFound
		}
		return nil, fmt.Errorf("failed to find auth provider: %w", err)
	}

	// Decrypt secrets
	if err := r.decryptSecrets(&m); err != nil {
		return nil, fmt.Errorf("failed to decrypt auth provider secrets: %w", err)
	}

	return fromModelToEntityAuthProvider(&m), nil
}

func (r *authProviderRepository) GetAll(ctx context.Context, projectIds []string) ([]entity.AuthProvider, error) {
	md := models.GetAuthProviderModel()
	filter := bson.D{{Key: md.ProjectIdKey, Value: bson.D{{Key: "$in", Value: projectIds}}}}

	cursor, err := r.db.Find(ctx, md, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find auth providers: %w", err)
	}
	defer cursor.Close(ctx)

	var modelsList []models.AuthProvider
	if err := cursor.All(ctx, &modelsList); err != nil {
		return nil, fmt.Errorf("failed to decode auth providers: %w", err)
	}

	providers := make([]entity.AuthProvider, len(modelsList))
	for i := range modelsList {
		if err := r.decryptSecrets(&modelsList[i]); err != nil {
			return nil, fmt.Errorf("failed to decrypt auth provider secrets at index %d: %w", i, err)
		}
		providers[i] = *fromModelToEntityAuthProvider(&modelsList[i])
	}

	return providers, nil
}

// Helpers

func (r *authProviderRepository) encryptSecrets(m *models.AuthProvider) error {
	for i := range m.Params {
		if m.Params[i].IsSecret && m.Params[i].Value != "" {
			encrypted, err := r.enc.Encrypt(m.Params[i].Value)
			if err != nil {
				return err
			}
			m.Params[i].Value = encrypted
		}
	}
	return nil
}

func (r *authProviderRepository) decryptSecrets(m *models.AuthProvider) error {
	for i := range m.Params {
		if m.Params[i].IsSecret && m.Params[i].Value != "" {
			decrypted, err := r.enc.Decrypt(m.Params[i].Value)
			if err != nil {
				return err
			}
			m.Params[i].Value = decrypted
		}
	}
	return nil
}

func fromEntityAuthProviderToModel(e entity.AuthProvider) models.AuthProvider {
	params := make([]models.AuthProviderParam, len(e.Params))
	for i, p := range e.Params {
		params[i] = models.AuthProviderParam{
			Label:    p.Label,
			Value:    p.Value,
			Key:      p.Key,
			IsSecret: p.IsSecret,
		}
	}

	return models.AuthProvider{
		Id:        e.Id,
		Name:      e.Name,
		Icon:      e.Icon,
		Provider:  models.AuthProviderType(e.Provider),
		Params:    params,
		ProjectId: e.ProjectId,
		Enabled:   e.Enabled,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		CreatedBy: e.CreatedBy,
		UpdatedBy: e.UpdatedBy,
	}
}

func fromModelToEntityAuthProvider(m *models.AuthProvider) *entity.AuthProvider {
	if m == nil {
		return nil
	}

	params := make([]entity.AuthProviderParam, len(m.Params))
	for i, p := range m.Params {
		params[i] = entity.AuthProviderParam{
			Label:    p.Label,
			Value:    p.Value,
			Key:      p.Key,
			IsSecret: p.IsSecret,
		}
	}

	return &entity.AuthProvider{
		Id:        m.Id,
		Name:      m.Name,
		Icon:      m.Icon,
		Provider:  entity.AuthProviderType(m.Provider),
		Params:    params,
		ProjectId: m.ProjectId,
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		CreatedBy: m.CreatedBy,
		UpdatedBy: m.UpdatedBy,
	}
}
