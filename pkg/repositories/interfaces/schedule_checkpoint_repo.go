package interfaces

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

//go:generate mockery -name=ScheduleCheckPointRepoInterface -output=../mocks -case=underscore

type ScheduleCheckPointRepoInterface interface {
	// Create or Update the checkpoint in the database store.
	Update(ctx context.Context, input models.ScheduleCheckPoint) error

	// Get a schedulable entity from the database store using the schedulable entity id.
	Get(ctx context.Context) (models.ScheduleCheckPoint, error)
}
