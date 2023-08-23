package common

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteadmin/pkg/errors"
)

func TestSortParameter_Empty(t *testing.T) {
	sortParameter, err := NewSortParameter(nil, sets.NewString())

	assert.NoError(t, err)
	assert.Nil(t, sortParameter)
}

func TestSortParameter_InvalidSortKey(t *testing.T) {
	expected := errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid sort_key: wrong")

	_, err := NewSortParameter(&admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "wrong",
	}, sets.NewString("name"))

	assert.Equal(t, expected, err)
}

func TestSortParameter_Ascending(t *testing.T) {
	sortParameter, err := NewSortParameter(&admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "name",
	}, sets.NewString("name"))

	assert.NoError(t, err)
	assert.Equal(t, "name asc", sortParameter[0].GetGormOrderExpr())
}

func TestSortParameter_Descending(t *testing.T) {
	sortParameter, err := NewSortParameter(&admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "project",
	}, sets.NewString("project"))

	assert.NoError(t, err)
	assert.Equal(t, "project desc", sortParameter[0].GetGormOrderExpr())
}
