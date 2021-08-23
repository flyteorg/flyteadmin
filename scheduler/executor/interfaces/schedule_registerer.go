package interfaces

import "context"

type ScheduleRegisterer interface {
	RunRegisterer(ctx context.Context)
}
