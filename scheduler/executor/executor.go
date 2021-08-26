package executor

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
)

type Executor interface {
	Execute(ctx context.Context, scheduledTime time.Time, s models.SchedulableEntity) error
}
