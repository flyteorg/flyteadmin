package implementations

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/events/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
)

type nodeExecutionEventWriter struct {
	db     repositories.RepositoryInterface
	events chan admin.NodeExecutionEventRequest
}

func (w *nodeExecutionEventWriter) Write(event admin.NodeExecutionEventRequest) {
	w.events <- event
}

func (w *nodeExecutionEventWriter) Run() {
	for event := range w.events {
		eventModel, err := transformers.CreateNodeExecutionEventModel(event)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to transform event [%+v] to database model with err [%+v]", event, err)
			continue
		}
		err = w.db.NodeExecutionEventRepo().Create(context.TODO(), *eventModel)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to write event [%+v] to database with err [%+v]", event, err)
		}
	}
}

func NewNodeExecutionEventWriter(db repositories.RepositoryInterface) interfaces.NodeExecutionEventWriter {
	return &nodeExecutionEventWriter{
		db:     db,
		events: make(chan admin.NodeExecutionEventRequest),
	}
}
