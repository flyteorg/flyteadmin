package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with signal models.
type SignalRepoInterface interface {
	// Inserts a signal model into the database store.
	GetOrCreate(ctx context.Context, input *models.Signal) error
	// Returns a matching signal if it exists.
	//Get(ctx context.Context, input GetSignalInput) (models.Signal, error)
	// Updates a signal model in the database store.
	Update(ctx context.Context, input models.Signal) error
}

type GetSignalInput struct {
	SignalID core.SignalIdentifier
}
