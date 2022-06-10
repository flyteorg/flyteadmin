package interfaces

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Defines the interface for interacting with signal models.
type SignalRepoInterface interface {
	// GetOrCreate inserts a signal model into the database store or returns one if it already exists.
	GetOrCreate(ctx context.Context, input *models.Signal) error
	// List a matching signal if it exists.
	List(ctx context.Context, input models.Signal) ([]*models.Signal, error)
	// Update updates the signal value in the database store.
	Update(ctx context.Context, input models.Signal) error
}

type GetSignalInput struct {
	SignalID core.SignalIdentifier
}
