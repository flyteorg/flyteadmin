package gormimpl

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

// SchedulableEntityRepo Implementation of SchedulableEntityRepoInterface.
type SchedulableEntityRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *SchedulableEntityRepo) GetAllActive(ctx context.Context) (models.SchedulableEntityCollectionOutput, error) {
	var schedulableEntities models.SchedulableEntityCollectionOutput
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Take(&schedulableEntities.Entities)
	timer.Stop()

	if tx.Error != nil {
		return models.SchedulableEntityCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.SchedulableEntityCollectionOutput{},
			fmt.Errorf("no active schedulable entities found")
	}
	return schedulableEntities, nil
}

func (r *SchedulableEntityRepo) UpdateLastExecution(ctx context.Context, input models.SchedulableEntity) error {

	timer := r.metrics.GetDuration.Start()
	// Update lastExecutionTime in the DB
	tx := r.db.Model(&models.SchedulableEntity{}).Where("project = ? AND domain=? AND name=? AND version=? ",
		input.Project, input.Domain, input.Name, input.Version).Update("last_execution_time", input.LastExecutionTime)

	timer.Stop()

	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *SchedulableEntityRepo) Create(ctx context.Context, input models.SchedulableEntity) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *SchedulableEntityRepo) Get(ctx context.Context, ID models.SchedulableEntityKey) (models.SchedulableEntity, error) {
	var schedulableEntity models.SchedulableEntity
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: ID.Project,
			Domain: ID.Domain,
			Name: ID.Name,
			Version: ID.Version,
		},
	}).Take(&schedulableEntity)
	timer.Stop()

	if tx.Error != nil {
		return models.SchedulableEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.SchedulableEntity{},
			errors.GetMissingEntityError("schedulable entity",  &core.Identifier{
				Project: ID.Project,
				Domain:  ID.Domain,
				Name:    ID.Name,
				Version: ID.Version,
			})
	}
	return schedulableEntity, nil
}

// NewSchedulableEntityRepo Returns an instance of SchedulableEntityRepoInterface
func NewSchedulableEntityRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.SchedulableEntityRepoInterface {
	metrics := newMetrics(scope)
	return &SchedulableEntityRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}

