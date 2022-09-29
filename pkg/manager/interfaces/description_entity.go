package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// DescriptionEntityInterface for managing DescriptionEntity
type DescriptionEntityInterface interface {
	CreateDescriptionEntity(ctx context.Context, request admin.DescriptionEntityCreateRequest) (*admin.DescriptionEntityCreateResponse, error)
	GetDescriptionEntity(ctx context.Context, request admin.ObjectGetRequest) (*admin.DescriptionEntity, error)
	ListDescriptionEntity(ctx context.Context, request admin.DescriptionEntityListRequest) (*admin.DescriptionEntityList, error)
}
