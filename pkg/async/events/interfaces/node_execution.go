package interfaces

import (
	"context"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name=NodeExecutionEventWriter -output=../mocks -case=underscore

type NodeExecutionEventWriter interface {
	Run(ctx context.Context)
	Write(nodeExecutionEvent admin.NodeExecutionEventRequest)
}
