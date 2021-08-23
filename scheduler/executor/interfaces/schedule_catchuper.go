package interfaces

import "context"

type ScheduleCatchuper interface {
	RunCatchuper(ctx context.Context) chan error
}
