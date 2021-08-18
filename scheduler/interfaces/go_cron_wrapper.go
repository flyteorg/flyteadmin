package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"time"
)


type RegisterFuncRef func(ctx context.Context, schedule models.SchedulableEntity, scheduledTime time.Time)

// GoCronWrapper Wrapper interface to the gogf framework which is used for scheduler. It also locks in schedule time.
type GoCronWrapper interface {
	// Register add and remove cron or fixed rate schedules from scheduler
	Register(ctx context.Context, s models.SchedulableEntity, funcRef RegisterFuncRef) error

	// DeRegister remove cron or fixed rate schedules from scheduler
	DeRegister(ctx context.Context, s models.SchedulableEntity)

	// GetScheduledTime returns the next scheduleTime from marker fromTime for the given schedule s
	GetScheduledTime(s models.SchedulableEntity, fromTime time.Time) (time.Time, error)

	// GetCatchUpTimes returns a slice of the all the schedules between from inclusive and to time exclusive for given schedule
	GetCatchUpTimes(schedule models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error)
}
