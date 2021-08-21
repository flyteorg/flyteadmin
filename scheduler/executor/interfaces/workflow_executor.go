package interfaces

import "context"

// WorkflowExecutor Handles reading the new workflow schedules and creating executions for them.
type WorkflowExecutor interface {
	Run(ctx context.Context) error
}
