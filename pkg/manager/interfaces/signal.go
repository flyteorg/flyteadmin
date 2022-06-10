package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing Flyte Signals
type SignalInterface interface {
	GetOrCreateSignal(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error)
	ListSignals(ctx context.Context, request admin.SignalListRequest) ([]*admin.Signal, error)
	SetSignal(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error)
}