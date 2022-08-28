package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

//go:generate mockery -name=NodeExecutionEventRepoInterface -output=../mocks -case=underscore

type NodeExecutionEventRepoInterface interface {
	// Inserts a node execution event into the database store.
	Create(ctx context.Context, input models.NodeExecutionEvent) error
	// Deletes a node execution event from the database store.
	Delete(ctx context.Context, input models.NodeExecutionEvent) error
	// List returns node executions matching query parameters. A limit must be provided for the results page size.
	// TODO Unit Test
	List(ctx context.Context, input ListResourceInput) (NodeExecutionEventCollectionOutput, error)
}
