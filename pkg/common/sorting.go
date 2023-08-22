package common

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/errors"
)

const gormDescending = "%s desc"
const gormAscending = "%s asc"

type SortParameter interface {
	GetGormOrderExpr() string
}

type sortParamImpl struct {
	gormOrderExpression string
}

func (s *sortParamImpl) GetGormOrderExpr() string {
	return s.gormOrderExpression
}

func NewSortParameter(sort *admin.Sort, allowed sets.String) ([]SortParameter, error) {
	if !allowed.Has(sort.Key) {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid sort_key: %s", sort.Key)
	}

	var gormOrderExpression string
	switch sort.Direction {
	case admin.Sort_DESCENDING:
		gormOrderExpression = fmt.Sprintf(gormDescending, sort.Key)
	case admin.Sort_ASCENDING:
		gormOrderExpression = fmt.Sprintf(gormAscending, sort.Key)
	default:
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid sort order specified: %v", sort)
	}

	return []SortParameter{&sortParamImpl{gormOrderExpression: gormOrderExpression}}, nil
}

func NewSortParameters(request *admin.ResourceListRequest, allowed sets.String) ([]SortParameter, error) {
	if len(request.SortKeys) > 0 && request.SortBy != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "cannot specify both sort_keys and sort_by")
	}

	if request.SortBy != nil {
		request.SortKeys = append(request.SortKeys, request.SortBy)
	}

	sortParams := make([]SortParameter, 0, len(request.SortKeys))
	for _, sortKey := range request.SortKeys {
		params, err := NewSortParameter(sortKey, allowed)
		if err != nil {
			return sortParams, err
		}
		sortParams = append(sortParams, params...)
	}

	return sortParams, nil
}
