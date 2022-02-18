package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with task execution models.
type TaskExecutionRepoInterface interface {
	// Create inserts a task execution model into the database store.
	Create(ctx context.Context, input models.TaskExecution) error
	// Update an existing task execution in the database store with all non-empty fields in the input.
	// Filters ensure only a matching, existing execution will be updated.
	Update(ctx context.Context, execution models.TaskExecution, filters []common.InlineFilter) error
	// Get return a matching execution if it exists.
	Get(ctx context.Context, input GetTaskExecutionInput) (models.TaskExecution, error)
	// List returns task executions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (TaskExecutionCollectionOutput, error)
}

type GetTaskExecutionInput struct {
	TaskExecutionID core.TaskExecutionIdentifier
}

// TaskExecutionCollectionOutput is a response format for a query on task executions.
type TaskExecutionCollectionOutput struct {
	TaskExecutions []models.TaskExecution
}
