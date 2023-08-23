package impl

import (
	"strings"

	"github.com/flyteorg/flyteadmin/pkg/common"
)

func sortParamsSQL(params []common.SortParameter) string {
	sqls := make([]string, len(params))
	for i, param := range params {
		sqls[i] = param.GetGormOrderExpr()
	}
	return strings.Join(sqls, ", ")
}
