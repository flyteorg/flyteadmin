package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

//go:generate mockery -name=SchedulableEntityRepoInterface -output=../mocks -case=underscore

// ScheduleEntitiesSnapShotRepoInterface : An Interface for interacting with the snapshot of schedulable entities in the database
type ScheduleEntitiesSnapShotRepoInterface interface {

	// Create/ Update the snapshot in the  database store
	CreateSnapShot(ctx context.Context, input models.ScheduleEntitiesSnapshot) error

	// Get the latest snapshot from the database store.
	GetLatestSnapShot(ctx context.Context) (models.ScheduleEntitiesSnapshot, error)
}
