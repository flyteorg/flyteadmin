package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Interface for managing Flyte admin version
type VersionInterface interface {
	GetVersion(ctx context.Context) (*admin.Version, error)
}
