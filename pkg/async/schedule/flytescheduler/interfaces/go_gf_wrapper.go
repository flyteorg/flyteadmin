package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"time"
)

// GoGFWrapper Wrapper interface to the gogf framework which is used for scheduler
type GoGFWrapper interface {
	// Register add and remove cron or fixed rate schedules from gogf
	Register(ctx context.Context, s models.SchedulableEntity, funcRef func()) error
	// GetScheduledTime returns the next scheduleTime from marker fromTime for the given schedule s
	GetScheduledTime(s models.SchedulableEntity, fromTime time.Time) (time.Time, error)
	// GetCatchUpTimes returns a slice of the all the schedules between from inclusive and to time exclusive for given schedule
	GetCatchUpTimes(schedule models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error)
	// GetScheduleName gets the unique name for the given schedule used in gogf
	GetScheduleName(schedule models.SchedulableEntity) string
}
