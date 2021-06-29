package interfaces

import "context"

// Handles responding to scheduled workflow execution events and creating executions.
type WorkflowExecutor interface {
	Run(ctx context.Context)
	Stop(ctx context.Context) error
}
