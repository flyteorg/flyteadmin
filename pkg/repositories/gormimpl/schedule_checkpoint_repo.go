package gormimpl

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/jinzhu/gorm"
)

// SchedulableEntityRepo Implementation of SchedulableEntityRepoInterface.
type ScheduleCheckPointRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *ScheduleCheckPointRepo) Update(ctx context.Context, input models.ScheduleCheckPoint) error {
	timer := r.metrics.GetDuration.Start()
	// Update lastExecutionTime in the DB

	tx := r.db.Model(&models.ScheduleCheckPoint{}).Save(input)
	timer.Stop()

	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}


func (r *ScheduleCheckPointRepo) Get(ctx context.Context) (models.ScheduleCheckPoint, error) {
	var scheduleCheckPoint models.ScheduleCheckPoint
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Order("check_point_time desc").Take(&scheduleCheckPoint)
	timer.Stop()

	if tx.Error != nil {
		return models.ScheduleCheckPoint{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RecordNotFound() {
		return models.ScheduleCheckPoint{}, nil
	}
	return scheduleCheckPoint, nil
}

// NewScheduleCheckPointRepo Returns an instance of ScheduleCheckPointRepoInterface
func NewScheduleCheckPointRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.ScheduleCheckPointRepoInterface {
	metrics := newMetrics(scope)
	return &ScheduleCheckPointRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}

