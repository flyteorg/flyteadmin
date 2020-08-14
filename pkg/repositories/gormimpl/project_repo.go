package gormimpl

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/jinzhu/gorm"

	"github.com/lyft/flyteadmin/pkg/common"
	flyteAdminErrors "github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
)

type ProjectRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ProjectRepo) Create(ctx context.Context, project models.Project) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Create(&project)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *ProjectRepo) Get(ctx context.Context, projectID string) (models.Project, error) {
	var project models.Project
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.Project{
		Identifier: projectID,
	}).First(&project)
	timer.Stop()
	if tx.Error != nil {
		return models.Project{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.Project{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "project [%s] not found", projectID)
	}
	return project, nil
}

func (r *ProjectRepo) ListAll(ctx context.Context, sortParameter common.SortParameter) ([]models.Project, error) {
	var projects []models.Project
	var tx = r.db
	if sortParameter != nil {
		tx = tx.Order(sortParameter.GetGormOrderExpr())
	}
	tx = tx.Find(&projects)
	if tx.Error != nil {
		return nil, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return projects, nil
}

func NewProjectRepo(db *gorm.DB, errorTransformer errors.ErrorTransformer,
	scope promutils.Scope) interfaces.ProjectRepoInterface {
	metrics := newMetrics(scope)
	return &ProjectRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}

func (r *ProjectRepo) UpdateProject(ctx context.Context, updatedProject models.Project) (error) {
	var project models.Project

	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.Project{
		Identifier: updatedProject.Identifier,
	}).First(&project)
	timer.Stop()

	// Error handling and checking for the result of the database query.
	if tx.Error != nil {
		r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	if tx.RecordNotFound() {
		flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "project [%s] not found", updatedProject.Identifier)
	}

	// Modify below fields if not null in the updatedProject.
	if updatedProject.Description != "" {
		project.Description = updatedProject.Description;
	}

	if len(updatedProject.Labels) > 0 {
		project.Labels = updatedProject.Labels;
	}

	// Use gorm client to update the two fields that are changed.
	r.db.Model(&project).Update("Description", "Labels")

	return nil
}
