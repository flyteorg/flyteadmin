package impl

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
)

func NewMissingEntityError(entity string) error {
	return errors.NewFlyteAdminErrorf(codes.NotFound, "Failed to find [%s]", entity)
}

var descCreatedAtSortParam = admin.Sort{
	Direction: admin.Sort_DESCENDING,
	Key:       "created_at",
}

var descCreatedAtSortDBParam, _ = common.NewSortParameter(descCreatedAtSortParam)
