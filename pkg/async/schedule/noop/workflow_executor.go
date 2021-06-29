package noop

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
)

type workflowExecutor struct{}

func (w *workflowExecutor) Run(ctx context.Context) {}

func (w *workflowExecutor) Stop(ctx context.Context) error {
	return nil
}

func NewWorkflowExecutor() interfaces.WorkflowExecutor {
	return &workflowExecutor{}
}
