package interfaces

import "context"

type ScheduleCheckPointer interface {
	RunCheckPointer(ctx context.Context)
	CheckPointState(ctx context.Context)
	ReadCheckPoint(ctx context.Context) Snapshoter
}
