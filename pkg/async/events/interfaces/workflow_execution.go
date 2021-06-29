package interfaces

import (
	"context"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name=WorkflowExecutionEventWriter -output=../mocks -case=underscore

type WorkflowExecutionEventWriter interface {
	Run(ctx context.Context)
	Write(workflowExecutionEvent admin.WorkflowExecutionEventRequest)
}
