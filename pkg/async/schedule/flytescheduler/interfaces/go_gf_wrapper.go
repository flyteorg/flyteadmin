package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"time"
)

type GoGFWrapper interface {
	Register(ctx context.Context, s models.SchedulableEntity, funcRef func()) error
	GetScheduledTime(s models.SchedulableEntity, fromTime time.Time) (time.Time, error)
	GetCatchUpTimes(schedule models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error)
	GetScheduleName(schedule models.SchedulableEntity) string
}

