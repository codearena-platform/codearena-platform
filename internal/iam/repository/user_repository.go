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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	GetById(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string, projectId string) (*entity.User, error)
	GetByPhone(ctx context.Context, phone string, projectId string) (*entity.User, error)
	GetAll(ctx context.Context, query sdk.UserQuery) (*sdk.UserList, error)
	RemoveResourceFromAll(ctx context.Context, resourceKey string) error
}

type userRepository struct {
	db db.DB
}

func NewUserRepository(db db.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	d := fromEntityToModel(*user)
	md := models.GetUserModel()
	_, err := r.db.InsertOne(ctx, md, d)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	d := fromEntityToModel(*user)
	md := models.GetUserModel()
	_, err := r.db.UpdateOne(ctx, md, bson.D{{Key: md.IdKey, Value: user.Id}}, bson.D{{Key: "$set", Value: d}})
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

func (r *userRepository) GetById(ctx context.Context, id string) (*entity.User, error) {
	md := models.GetUserModel()
	var usr models.User
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.IdKey, Value: id}}).Decode(&usr)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	return fromModelToEntity(&usr), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string, projectId string) (*entity.User, error) {
	md := models.GetUserModel()
	var usr models.User
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.EmailKey, Value: email}, {Key: md.ProjectIDKey, Value: projectId}}).Decode(&usr)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	return fromModelToEntity(&usr), nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string, projectId string) (*entity.User, error) {
	md := models.GetUserModel()
	var usr models.User
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.PhoneKey, Value: phone}, {Key: md.ProjectIDKey, Value: projectId}}).Decode(&usr)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	return fromModelToEntity(&usr), nil
}

func (r *userRepository) GetAll(ctx context.Context, query sdk.UserQuery) (*sdk.UserList, error) {
	md := models.GetUserModel()
	filter := bson.A{}
	var users []models.User
	if query.SearchQuery != "" {
		filter = append(filter,
			bson.D{{Key: md.NameKey, Value: bson.D{{Key: "$regex", Value: query.SearchQuery}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: md.EmailKey, Value: bson.D{{Key: "$regex", Value: query.SearchQuery}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: md.PhoneKey, Value: bson.D{{Key: "$regex", Value: query.SearchQuery}, {Key: "$options", Value: "i"}}}},
		)
	}
	cond := bson.D{{Key: md.EnabledKey, Value: true}, {Key: md.ProjectIDKey, Value: bson.D{{Key: "$in", Value: query.ProjectIds}}}}

	if len(query.RoleId) > 0 {
		cond = append(cond, bson.E{Key: fmt.Sprintf("%s.%s", md.RolesIdKey, query.RoleId), Value: bson.D{{Key: "$exists", Value: true}}})
	}

	if len(filter) > 0 {
		cond = append(cond, bson.E{Key: "$or", Value: filter})
	}

	total, err := r.db.CountDocuments(ctx, md, cond)
	if err != nil {
		return nil, fmt.Errorf("error counting users: %w", err)
	}

	opts := options.Find().
		SetSkip(query.Skip).
		SetLimit(query.Limit)

	cursor, err := r.db.Find(ctx, md, cond, opts)
	if err != nil {
		return nil, fmt.Errorf("error finding all users: %w", err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, fmt.Errorf("error reading all users: %w", err)
	}

	entityUsers := make([]entity.User, len(users))
	for i := range users {
		entityUsers[i] = *fromModelToEntity(&users[i])
	}

	return &sdk.UserList{
		Users: entityUsers,
		Total: total,
		Skip:  query.Skip,
		Limit: query.Limit,
	}, nil
}

func (r *userRepository) RemoveResourceFromAll(ctx context.Context, resourceKey string) error {
	md := models.GetUserModel()
	filter := bson.D{{Key: fmt.Sprintf("%s.%s", md.ResourcesKey, resourceKey), Value: bson.D{{Key: "$exists", Value: true}}}}
	update := bson.D{{Key: "$unset", Value: bson.D{{Key: fmt.Sprintf("%s.%s", md.ResourcesKey, resourceKey), Value: ""}}}}

	_, err := r.db.UpdateMany(ctx, md, filter, update)
	if err != nil {
		return fmt.Errorf("error removing resource from all users: %w", err)
	}
	return nil
}

// Helpers

func fromEntityToModel(user entity.User) models.User {
	return models.User{
		Id:             user.Id,
		Email:          user.Email,
		Phone:          user.Phone,
		Name:           user.Name,
		ProjectId:      user.ProjectId,
		Enabled:        user.Enabled,
		Expiry:         user.Expiry,
		ProfilePic:     user.ProfilePic,
		LinkedClientId: user.LinkedClientId,
		Roles:          fromEntityUserRoleMapToModel(user.Roles),
		Resources:      fromEntityUserResourceMapToModel(user.Resources),
		Policies:       fromEntityUserPoliciesToModel(user.Policies),
		CreatedAt:      user.CreatedAt,
		CreatedBy:      user.CreatedBy,
		UpdatedAt:      user.UpdatedAt,
		UpdatedBy:      user.UpdatedBy,
	}
}

func fromModelToEntity(user *models.User) *entity.User {
	return &entity.User{
		Id:             user.Id,
		Email:          user.Email,
		Phone:          user.Phone,
		Name:           user.Name,
		ProfilePic:     user.ProfilePic,
		ProjectId:      user.ProjectId,
		Expiry:         user.Expiry,
		Enabled:        user.Enabled,
		LinkedClientId: user.LinkedClientId,
		Roles:          fromModelUserRoleMapToEntity(user.Roles),
		Resources:      fromModelUserResourceMapToEntity(user.Resources),
		Policies:       fromModelUserPoliciesToEntity(user.Policies),
		CreatedAt:      user.CreatedAt,
		CreatedBy:      user.CreatedBy,
		UpdatedAt:      user.UpdatedAt,
		UpdatedBy:      user.UpdatedBy,
	}
}

func fromEntityUserPoliciesToModel(policies map[string]entity.UserPolicy) map[string]models.UserPolicy {
	result := map[string]models.UserPolicy{}
	for key, policy := range policies {
		result[key] = models.UserPolicy{
			Name:    policy.Name,
			Mapping: fromEntityUserPolicyMappingToModel(policy.Mapping),
		}
	}
	return result
}

func fromEntityUserPolicyMappingToModel(mapping entity.UserPolicyMapping) models.UserPolicyMapping {
	arguments := map[string]models.UserPolicyMappingValue{}
	for key, argument := range mapping.Arguments {
		arguments[key] = models.UserPolicyMappingValue{
			Static: argument.Static,
		}
	}
	return models.UserPolicyMapping{Arguments: arguments}
}

func fromModelUserPoliciesToEntity(policies map[string]models.UserPolicy) map[string]entity.UserPolicy {
	result := map[string]entity.UserPolicy{}
	for key, policy := range policies {
		result[key] = entity.UserPolicy{
			Name:    policy.Name,
			Mapping: fromModelUserPolicyMappingToEntity(policy.Mapping),
		}
	}
	return result
}

func fromModelUserPolicyMappingToEntity(mapping models.UserPolicyMapping) entity.UserPolicyMapping {
	arguments := map[string]entity.UserPolicyMappingValue{}
	for key, argument := range mapping.Arguments {
		arguments[key] = entity.UserPolicyMappingValue{
			Static: argument.Static,
		}
	}
	return entity.UserPolicyMapping{Arguments: arguments}
}

func fromEntityUserRoleMapToModel(roles map[string]entity.UserRole) map[string]models.UserRoles {
	userRoles := make(map[string]models.UserRoles)
	for key, role := range roles {
		userRoles[key] = models.UserRoles{
			Name: role.Name,
			Id:   role.Id,
		}
	}
	return userRoles
}

func fromModelUserRoleMapToEntity(roles map[string]models.UserRoles) map[string]entity.UserRole {
	userRoles := make(map[string]entity.UserRole)
	for key, role := range roles {
		userRoles[key] = entity.UserRole{
			Name: role.Name,
			Id:   role.Id,
		}
	}
	return userRoles
}

func fromEntityUserResourceMapToModel(resources map[string]entity.UserResource) map[string]models.UserResource {
	userResources := make(map[string]models.UserResource)
	for key, res := range resources {
		userResources[key] = models.UserResource{
			PolicyIds: res.PolicyIds,
			RoleIds:   res.RoleIds,
			Key:       res.Key,
			Name:      res.Name,
		}
	}
	return userResources
}

func fromModelUserResourceMapToEntity(resources map[string]models.UserResource) map[string]entity.UserResource {
	userResources := make(map[string]entity.UserResource)
	for key, res := range resources {
		userResources[key] = entity.UserResource{
			PolicyIds: res.PolicyIds,
			RoleIds:   res.RoleIds,
			Key:       res.Key,
			Name:      res.Name,
		}
	}
	return userResources
}
