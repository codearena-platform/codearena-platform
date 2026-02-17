package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	ErrRoleNotFound = errors.New("role not found")
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	GetById(ctx context.Context, id string) (*entity.Role, error)
	GetAll(ctx context.Context, query sdk.RoleQuery) (*sdk.RoleList, error)
	RemoveResourceFromAll(ctx context.Context, resourceKey string) error
}

type roleRepository struct {
	db db.DB
}

func NewRoleRepository(db db.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	d := fromEntityRoleToModel(*role)
	md := models.GetRoleModel()
	_, err := r.db.InsertOne(ctx, md, d)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	d := fromEntityRoleToModel(*role)
	md := models.GetRoleModel()
	result, err := r.db.UpdateOne(
		ctx,
		md,
		bson.D{{Key: md.IdKey, Value: role.Id}},
		bson.D{{Key: "$set", Value: d}},
	)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	if result.ModifiedCount == 0 {
		return ErrRoleNotFound
	}
	return nil
}

func (r *roleRepository) GetById(ctx context.Context, id string) (*entity.Role, error) {
	md := models.GetRoleModel()
	var roleMod models.Role
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.IdKey, Value: id}}).Decode(&roleMod)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to find role: %w", err)
	}
	return fromModelToEntityRole(&roleMod), nil
}

func (r *roleRepository) GetAll(ctx context.Context, query sdk.RoleQuery) (*sdk.RoleList, error) {
	md := models.GetRoleModel()
	filter := bson.A{}

	if query.SearchQuery != "" {
		filter = append(filter, bson.D{{Key: md.NameKey, Value: primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", query.SearchQuery), Options: "i"}}})
		filter = append(filter, bson.D{{Key: md.DescriptionKey, Value: primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", query.SearchQuery), Options: "i"}}})
	}

	cond := bson.D{{Key: md.EnabledKey, Value: true}, {Key: md.ProjectIdKey, Value: bson.D{{Key: "$in", Value: query.ProjectIds}}}}

	if len(filter) > 0 {
		cond = bson.D{{Key: "$or", Value: filter}}
	}

	total, err := r.db.CountDocuments(ctx, md, cond)
	if err != nil {
		return nil, fmt.Errorf("error counting roles: %w", err)
	}

	opts := options.Find().
		SetSkip(query.Skip).
		SetLimit(query.Limit)

	var roles []models.Role
	cursor, err := r.db.Find(ctx, md, cond, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &roles); err != nil {
		return nil, fmt.Errorf("failed to read roles: %w", err)
	}

	entityRoles := make([]entity.Role, len(roles))
	for i := range roles {
		entityRoles[i] = *fromModelToEntityRole(&roles[i])
	}

	return &sdk.RoleList{
		Roles: entityRoles,
		Total: total,
		Skip:  query.Skip,
		Limit: query.Limit,
	}, nil
}

func (r *roleRepository) RemoveResourceFromAll(ctx context.Context, resourceKey string) error {
	md := models.GetRoleModel()
	_, err := r.db.UpdateMany(ctx, md, bson.D{
		{Key: fmt.Sprintf("%s.%s", md.ResourcesKey, resourceKey), Value: bson.D{{Key: "$exists", Value: true}}},
	}, bson.D{{Key: "$unset", Value: bson.D{{Key: fmt.Sprintf("%s.%s", md.ResourcesKey, resourceKey), Value: ""}}}})
	if err != nil {
		return fmt.Errorf("failed to remove resource from roles: %w", err)
	}
	return nil
}

// Helpers

func fromEntityRoleToModel(role entity.Role) models.Role {
	createdAt := time.Time{}
	updatedAt := time.Time{}
	if role.CreatedAt != nil {
		createdAt = *role.CreatedAt
	}
	if role.UpdatedAt != nil {
		updatedAt = *role.UpdatedAt
	}

	return models.Role{
		Id:          role.Id,
		ProjectId:   role.ProjectId,
		Name:        role.Name,
		Description: role.Description,
		Resources:   fromEntityResourceMapToModel(role.Resources),
		CreatedAt:   createdAt,
		CreatedBy:   role.CreatedBy,
		UpdatedAt:   updatedAt,
		Enabled:     role.Enabled,
		UpdatedBy:   role.UpdatedBy,
	}
}

func fromModelToEntityRole(role *models.Role) *entity.Role {
	if role == nil {
		return nil
	}

	createdAt := role.CreatedAt
	updatedAt := role.UpdatedAt

	return &entity.Role{
		Id:          role.Id,
		ProjectId:   role.ProjectId,
		Name:        role.Name,
		Description: role.Description,
		Resources:   fromModelResourceMapToEntityRole(role.Resources),
		Enabled:     role.Enabled,
		CreatedAt:   &createdAt,
		CreatedBy:   role.CreatedBy,
		UpdatedAt:   &updatedAt,
		UpdatedBy:   role.UpdatedBy,
	}
}

func fromEntityResourceMapToModel(resources map[string]entity.Resources) map[string]models.Resources {
	result := make(map[string]models.Resources)
	for _, res := range resources {
		if res.Key != "" {
			result[res.Key] = models.Resources{
				Id:   res.Id,
				Key:  res.Key,
				Name: res.Name,
			}
		}
	}
	return result
}

func fromModelResourceMapToEntityRole(resources map[string]models.Resources) map[string]entity.Resources {
	result := make(map[string]entity.Resources)
	for key, res := range resources {
		if key != "" {
			result[key] = entity.Resources{
				Id:   res.Id,
				Key:  res.Key,
				Name: res.Name,
			}
		}
	}
	return result
}
