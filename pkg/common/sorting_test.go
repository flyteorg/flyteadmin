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

func TestSortParameters_SingleAndMultipleSortKeys(t *testing.T) {
	expected := errors.NewFlyteAdminErrorf(codes.InvalidArgument, "cannot specify both sort_keys and sort_by")

	_, err := NewSortParameters(&admin.ResourceListRequest{
		SortBy:   &admin.Sort{},
		SortKeys: []*admin.Sort{{}},
	}, sets.NewString())

	assert.Equal(t, expected, err)
}

func TestSortParameters_SingleSortKey(t *testing.T) {
	params, err := NewSortParameters(&admin.ResourceListRequest{
		SortBy: &admin.Sort{Key: "foo"},
	}, sets.NewString("foo"))

	assert.NoError(t, err)
	assert.Equal(t, "foo desc", params[0].GetGormOrderExpr())
}

func TestSortParameters_OK(t *testing.T) {
	params, err := NewSortParameters(&admin.ResourceListRequest{
		SortKeys: []*admin.Sort{
			{Key: "key"},
			{Key: "foo", Direction: admin.Sort_ASCENDING},
		},
	}, sets.NewString("key", "foo"))

	assert.NoError(t, err)
	if assert.Len(t, params, 2) {
		assert.Equal(t, "key desc", params[0].GetGormOrderExpr())
		assert.Equal(t, "foo asc", params[1].GetGormOrderExpr())
	}
}

func TestSortParameters_Invalid(t *testing.T) {
	expected := errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid sort_key: foo")

	_, err := NewSortParameters(&admin.ResourceListRequest{
		SortKeys: []*admin.Sort{
			{Key: "foo"},
		},
	}, sets.NewString("key"))

	assert.Equal(t, expected, err)
}
