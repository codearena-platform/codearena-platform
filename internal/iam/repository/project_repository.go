package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/codearena-platform/codearena-platform/internal/iam/db"
	"github.com/codearena-platform/codearena-platform/internal/iam/db/models"
	"github.com/codearena-platform/codearena-platform/internal/iam/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrProjectNotFound = errors.New("project not found")
)

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	Update(ctx context.Context, project *entity.Project) error
	GetById(ctx context.Context, id string) (*entity.Project, error)
	GetByName(ctx context.Context, name string) (*entity.Project, error)
	GetAll(ctx context.Context) ([]entity.Project, error)
}

type projectRepository struct {
	db db.DB
}

func NewProjectRepository(db db.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *entity.Project) error {
	d := fromEntityProjectToModel(*project)
	md := models.GetProjectModel()
	_, err := r.db.InsertOne(ctx, md, d)
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}
	return nil
}

func (r *projectRepository) Update(ctx context.Context, project *entity.Project) error {
	d := fromEntityProjectToModel(*project)
	md := models.GetProjectModel()
	_, err := r.db.UpdateOne(ctx, md, bson.D{{Key: md.IdKey, Value: project.Id}}, bson.D{{Key: "$set", Value: d}})
	if err != nil {
		return fmt.Errorf("error updating project: %w", err)
	}
	return nil
}

func (r *projectRepository) GetById(ctx context.Context, id string) (*entity.Project, error) {
	md := models.GetProjectModel()
	var proj models.Project
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.IdKey, Value: id}}).Decode(&proj)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("error finding project: %w", err)
	}
	return fromModelToEntityProject(&proj), nil
}

func (r *projectRepository) GetByName(ctx context.Context, name string) (*entity.Project, error) {
	md := models.GetProjectModel()
	var proj models.Project
	err := r.db.FindOne(ctx, md, bson.D{{Key: md.NameKey, Value: name}}).Decode(&proj)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("error finding project by name: %w", err)
	}
	return fromModelToEntityProject(&proj), nil
}

func (r *projectRepository) GetAll(ctx context.Context) ([]entity.Project, error) {
	md := models.GetProjectModel()
	var projects []models.Project
	cursor, err := r.db.Find(ctx, md, bson.D{{}})
	if err != nil {
		return nil, fmt.Errorf("error finding all projects: %w", err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &projects)
	if err != nil {
		return nil, fmt.Errorf("error reading all projects: %w", err)
	}

	entityProjects := make([]entity.Project, len(projects))
	for i := range projects {
		entityProjects[i] = *fromModelToEntityProject(&projects[i])
	}
	return entityProjects, nil
}

// Helpers

func fromEntityProjectToModel(proj entity.Project) models.Project {
	return models.Project{
		Id:          proj.Id,
		Name:        proj.Name,
		Description: proj.Description,
		Tags:        proj.Tags,
		CreatedAt:   proj.CreatedAt,
		CreatedBy:   proj.CreatedBy,
		UpdatedAt:   proj.UpdatedAt,
		UpdatedBy:   proj.UpdatedBy,
	}
}

func fromModelToEntityProject(proj *models.Project) *entity.Project {
	if proj == nil {
		return nil
	}
	return &entity.Project{
		Id:          proj.Id,
		Name:        proj.Name,
		Description: proj.Description,
		Tags:        proj.Tags,
		CreatedAt:   proj.CreatedAt,
		CreatedBy:   proj.CreatedBy,
		UpdatedAt:   proj.UpdatedAt,
		UpdatedBy:   proj.UpdatedBy,
	}
}
