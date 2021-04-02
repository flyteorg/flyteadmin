package implementations

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/events/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
)

type workflowExecutionEventWriter struct {
	db     repositories.RepositoryInterface
	events chan admin.WorkflowExecutionEventRequest
}

func (w *workflowExecutionEventWriter) Write(event admin.WorkflowExecutionEventRequest) {
	w.events <- event
}

func (w *workflowExecutionEventWriter) Run() {
	for event := range w.events {
		eventModel, err := transformers.CreateExecutionEventModel(event)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to transform event [%+v] to database model with err [%+v]", event, err)
			continue
		}
		err = w.db.ExecutionEventRepo().Create(context.TODO(), *eventModel)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to write event [%+v] to database with err [%+v]", event, err)
		}
	}
}

func NewWorkflowExecutionEventWriter(db repositories.RepositoryInterface) interfaces.WorkflowExecutionEventWriter {
	return &workflowExecutionEventWriter{
		db:     db,
		events: make(chan admin.WorkflowExecutionEventRequest),
	}
}
