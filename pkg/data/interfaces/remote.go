package interfaces

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/storage"
)

// Defines an interface for fetching pre-signed URLs.
type RemoteURLInterface interface {
	// TODO: Refactor for URI to be of type DataReference. We should package a FromString-like function in flytestdlib
	Get(ctx context.Context, uri string) (admin.UrlBlob, error)
}
