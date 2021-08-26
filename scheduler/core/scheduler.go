package core

import (
	"context"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteadmin/scheduler/snapshoter"
	"time"
)

type TimedFuncWithSchedule func(ctx context.Context, s models.SchedulableEntity, t time.Time) error

type Scheduler interface {
	ScheduleJob(ctx context.Context, s models.SchedulableEntity, f TimedFuncWithSchedule, lastT *time.Time) error
	DeScheduleJob(ctx context.Context, s models.SchedulableEntity)
	BootStrapSchedulesFromSnapShot(ctx context.Context, schedules []models.SchedulableEntity, snapshot executor.Snapshot)
	UpdateSchedules(ctx context.Context, s []models.SchedulableEntity)
	CalculateSnapshot(ctx context.Context) executor.Snapshot
	CatchupAll(ctx context.Context, until time.Time) bool
}
