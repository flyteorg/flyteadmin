package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"time"
)

type ScheduleExecutionFirer interface {
	Fire(ctx context.Context, scheduledTime time.Time, s models.SchedulableEntity) error
}
