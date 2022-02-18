package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

// Defines the interface for interacting with workflow execution models.
type ExecutionRepoInterface interface {
	// Create will insert a workflow execution model into the database store.
	Create(ctx context.Context, input models.Execution) error
	// Update applies only an existing execution model with all non-empty fields in the input.
	// Filters ensure only a matching, existing execution will be updated.
	Update(ctx context.Context, execution models.Execution, filters []common.InlineFilter) error
	// Get return a matching execution if it exists.
	Get(ctx context.Context, input Identifier) (models.Execution, error)
	// List returns executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (ExecutionCollectionOutput, error)
}

// ExecutionCollectionOutput is a response format for a query on workflows.
type ExecutionCollectionOutput struct {
	Executions []models.Execution
}
