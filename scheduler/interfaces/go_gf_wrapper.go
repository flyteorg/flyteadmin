package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"time"
)


type RegisterFuncRef func(ctx context.Context, schedule models.SchedulableEntity, scheduledTime time.Time)

// GoGFWrapper Wrapper interface to the gogf framework which is used for scheduler. It also locks in schedule time.
type GoGFWrapper interface {
	// Register add and remove cron or fixed rate schedules from gogf
	Register(ctx context.Context, s models.SchedulableEntity, asOfTime time.Time, funcRef RegisterFuncRef) error

	// DeRegister remove cron or fixed rate schedules from gogf
	DeRegister(ctx context.Context, s models.SchedulableEntity)

	// GetScheduledTime returns the next scheduleTime from marker fromTime for the given schedule s
	GetScheduledTime(s models.SchedulableEntity, fromTime time.Time) (time.Time, error)

	// GetCatchUpTimes returns a slice of the all the schedules between from inclusive and to time exclusive for given schedule
	GetCatchUpTimes(schedule models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error)
}
