package gormimpl

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

// ScheduleEntitiesSnapshotRepo Implementation of ScheduleEntitiesSnapshotRepoInterface.
type ScheduleEntitiesSnapshotRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

// TODO : always overwrite the exisiting snapshot instead of creating new rows
func (r *ScheduleEntitiesSnapshotRepo) CreateSnapShot(ctx context.Context, input models.ScheduleEntitiesSnapshot) error {
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *ScheduleEntitiesSnapshotRepo) GetLatestSnapShot(ctx context.Context) (models.ScheduleEntitiesSnapshot, error) {
	var schedulableEntitiesSnapshot models.ScheduleEntitiesSnapshot
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Last(&schedulableEntitiesSnapshot)
	timer.Stop()

	if tx.Error != nil {
		if tx.RecordNotFound() {
			return models.ScheduleEntitiesSnapshot{},
				fmt.Errorf("no snapshot exists in the db")
		}
		return models.ScheduleEntitiesSnapshot{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return schedulableEntitiesSnapshot, nil
}

// NewScheduleEntitiesSnapshotRepo Returns an instance of ScheduleEntitiesSnapshotRepoInterface
func NewScheduleEntitiesSnapshotRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.ScheduleEntitiesSnapShotRepoInterface {
	metrics := newMetrics(scope)
	return &ScheduleEntitiesSnapshotRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
