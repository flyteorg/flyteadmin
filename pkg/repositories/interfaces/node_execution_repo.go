package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with node execution models.
type NodeExecutionRepoInterface interface {
	// Create inserts a new node execution model and the first event that triggers it into the database store.
	Create(ctx context.Context, execution *models.NodeExecution) error
	// Update an existing node execution in the database store with all non-empty fields in the input.
	// Filters ensure only a matching, existing execution will be updated.
	Update(ctx context.Context, execution *models.NodeExecution, filters []common.InlineFilter) error
	// Get returns a matching execution if it exists.
	Get(ctx context.Context, input NodeExecutionResource) (models.NodeExecution, error)
	// List returns node executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (NodeExecutionCollectionOutput, error)
	// ListEvents returns node execution events matching query parameters. A limit must be provided for the results page size.
	ListEvents(ctx context.Context, input ListResourceInput) (NodeExecutionEventCollectionOutput, error)
	// Exists returns whether a matching execution  exists.
	Exists(ctx context.Context, input NodeExecutionResource) (bool, error)
}

type NodeExecutionResource struct {
	NodeExecutionIdentifier core.NodeExecutionIdentifier
}

// NodeExecutionCollectionOutput is a response format for a query on node executions.
type NodeExecutionCollectionOutput struct {
	NodeExecutions []models.NodeExecution
}

// NodeExecutionEventCollectionOutput is a response format for a query on node execution events.
type NodeExecutionEventCollectionOutput struct {
	NodeExecutionEvents []models.NodeExecutionEvent
}
