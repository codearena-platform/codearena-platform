package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"github.com/codearena-platform/codearena-platform/internal/iam/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
)

type ResourceRepository interface {
	Create(ctx context.Context, resource *entity.Resource) (string, error)
	Update(ctx context.Context, resource *entity.Resource) error
	GetById(ctx context.Context, id string) (*entity.Resource, error)
	GetAll(ctx context.Context, query sdk.ResourceQuery) (*sdk.ResourceList, error)
	Delete(ctx context.Context, id string) error
}

type resourceRepository struct {
	db db.DB
}

func NewResourceRepository(db db.DB) ResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) Create(ctx context.Context, resource *entity.Resource) (string, error) {
	d := fromEntityResourceToModel(*resource)
	md := models.GetResourceModel()
	_, err := r.db.InsertOne(ctx, md, d)
	if err != nil {
		return "", fmt.Errorf("error creating resource: %w", err)
	}
	return resource.ID, nil
}

func (r *resourceRepository) Update(ctx context.Context, resource *entity.Resource) error {
	d := fromEntityResourceToModel(*resource)
	md := models.GetResourceModel()
	_, err := r.db.UpdateOne(ctx, md, bson.D{{Key: md.IdKey, Value: resource.ID}}, bson.D{{Key: "$set", Value: d}})
	if err != nil {
		return fmt.Errorf("error updating resource: %w", err)
	}
	return nil
}

func (r *resourceRepository) GetById(ctx context.Context, id string) (*entity.Resource, error) {
	md := models.GetResourceModel()
	var resMod models.Resource
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.IdKey, Value: id}, {Key: md.EnabledKey, Value: true}}).Decode(&resMod)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("error finding resource: %w", err)
	}
	return fromModelToEntityResource(&resMod), nil
}

func (r *resourceRepository) GetAll(ctx context.Context, query sdk.ResourceQuery) (*sdk.ResourceList, error) {
	md := models.GetResourceModel()
	filter := bson.A{}

	if query.Name != "" {
		filter = append(filter, bson.D{{Key: md.NameKey, Value: primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", query.Name), Options: "i"}}})
	}
	if query.Description != "" {
		filter = append(filter, bson.D{{Key: md.DescriptionKey, Value: primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", query.Description), Options: "i"}}})
	}
	if query.Key != "" {
		filter = append(filter, bson.D{{Key: md.KeyKey, Value: primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", query.Key), Options: "i"}}})
	}

	cond := bson.D{{Key: md.EnabledKey, Value: true}, {Key: md.ProjectIdKey, Value: bson.D{{Key: "$in", Value: query.ProjectIds}}}}

	if len(filter) > 0 {
		cond = append(cond, bson.E{Key: "$or", Value: filter})
	}

	total, err := r.db.CountDocuments(ctx, md, cond)
	if err != nil {
		return nil, fmt.Errorf("error counting resources: %w", err)
	}

	opts := options.Find().
		SetSkip(query.Skip).
		SetLimit(query.Limit)

	var resources []models.Resource
	cursor, err := r.db.Find(ctx, md, cond, opts)
	if err != nil {
		return nil, fmt.Errorf("error finding resources: %w", err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &resources)
	if err != nil {
		return nil, fmt.Errorf("error reading resources: %w", err)
	}

	entityResources := make([]entity.Resource, len(resources))
	for i := range resources {
		entityResources[i] = *fromModelToEntityResource(&resources[i])
	}

	return &sdk.ResourceList{
		Resources: entityResources,
		Total:     total,
		Skip:      query.Skip,
		Limit:     query.Limit,
	}, nil
}

func (r *resourceRepository) Delete(ctx context.Context, id string) error {
	md := models.GetResourceModel()
	_, err := r.db.UpdateOne(ctx, md, bson.D{{Key: md.IdKey, Value: id}}, bson.D{{Key: "$set", Value: bson.D{{Key: md.EnabledKey, Value: false}}}})
	if err != nil {
		return fmt.Errorf("error deleting resource: %w", err)
	}
	return nil
}

// Helpers

func fromEntityResourceToModel(res entity.Resource) models.Resource {
	return models.Resource{
		ID:          res.ID,
		ProjectId:   res.ProjectId,
		Key:         res.Key,
		Name:        res.Name,
		Description: res.Description,
		Enabled:     res.Enabled,
		CreatedAt:   res.CreatedAt,
		CreatedBy:   res.CreatedBy,
		UpdatedAt:   res.UpdatedAt,
		UpdatedBy:   res.UpdatedBy,
		DeletedAt:   res.DeletedAt,
	}
}

func fromModelToEntityResource(res *models.Resource) *entity.Resource {
	if res == nil {
		return nil
	}
	return &entity.Resource{
		ID:          res.ID,
		ProjectId:   res.ProjectId,
		Key:         res.Key,
		Name:        res.Name,
		Description: res.Description,
		Enabled:     res.Enabled,
		CreatedAt:   res.CreatedAt,
		CreatedBy:   res.CreatedBy,
		UpdatedAt:   res.UpdatedAt,
		UpdatedBy:   res.UpdatedBy,
		DeletedAt:   res.DeletedAt,
	}
}
