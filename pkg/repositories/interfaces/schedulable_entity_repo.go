package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

//go:generate mockery -name=SchedulableEntityRepoInterface -output=../mocks -case=underscore

type SchedulableEntityRepoInterface interface {
	// Create a schedulable entity in the database store.
	Create(ctx context.Context, input models.SchedulableEntity) error

	// Get a schedulable entity from the database store using the schedulable entity id.
	Get(ctx context.Context, ID models.SchedulableEntityKey) (models.SchedulableEntity, error)

	// GetAllActive Gets all the active schedulable entities from the db
	GetAllActive(ctx context.Context) (models.SchedulableEntityCollectionOutput, error)

	// Update a schedulable entity with
	UpdateLastExecution(ctx context.Context, input models.SchedulableEntity) error

	// Delete a schedulable entity from the database store.
	// Delete(ctx context.Context, ID SchedulableEntityID) error
}
