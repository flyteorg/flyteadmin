package impl

import (
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/errors"
)

func NewMissingEntityError(entity string) error {
	return errors.NewFlyteAdminErrorf(codes.NotFound, "Failed to find [%s]", entity)
}
