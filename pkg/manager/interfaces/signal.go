package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing Flyte Signals
type SignalInterface interface {
	CreateSignal(ctx context.Context, request admin.SignalCreateRequest) (*admin.SignalCreateResponse, error)
	GetSignal(ctx context.Context, request admin.SignalGetRequest) (*admin.Signal, error)
}
