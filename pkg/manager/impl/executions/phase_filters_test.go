package executions

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestNewNotTerminalFilter(t *testing.T) {
	t.Run("workflow executions", func(t *testing.T) {
		filter, err := newNotTerminalFilter(common.Execution)
		assert.NoError(t, err)
		queryExpr, err := filter.GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase not in (?)")
		assert.EqualValues(t, queryExpr.Args, TerminalWorkflowExecutionPhases)
	})
	t.Run("node executions", func(t *testing.T) {
		filter, err := newNotTerminalFilter(common.NodeExecution)
		assert.NoError(t, err)
		queryExpr, err := filter.GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase not in (?)")
		assert.EqualValues(t, queryExpr.Args, TerminalNodeExecutionPhases)
	})
	t.Run("task executions", func(t *testing.T) {
		filter, err := newNotTerminalFilter(common.TaskExecution)
		assert.NoError(t, err)
		queryExpr, err := filter.GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase not in (?)")
		assert.EqualValues(t, queryExpr.Args, TerminalTaskExecutionPhases)
	})
}

func TestNewSchedulingFilter(t *testing.T) {
	filter, err := newSchedulingFilter()
	assert.NoError(t, err)
	queryExpr, err := filter.GetGormQueryExpr()
	assert.NoError(t, err)
	assert.Equal(t, queryExpr.Query, "phase in (?)")
	assert.EqualValues(t, queryExpr.Args, SchedulingWorkflowPhases)
}

func TestGetUpdateExecutionFilters(t *testing.T) {
	t.Run("queued", func(t *testing.T) {
		filters, err := GetUpdateExecutionFilters(core.WorkflowExecution_QUEUED)
		assert.NoError(t, err)
		assert.Len(t, filters, 2)

		queryExpr, err := filters[0].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase not in (?)")
		assert.EqualValues(t, queryExpr.Args, TerminalWorkflowExecutionPhases)

		queryExpr, err = filters[1].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase in (?)")
		assert.EqualValues(t, queryExpr.Args, SchedulingWorkflowPhases)
	})
	t.Run("running", func(t *testing.T) {
		filters, err := GetUpdateExecutionFilters(core.WorkflowExecution_RUNNING)
		assert.NoError(t, err)
		assert.Len(t, filters, 3)

		queryExpr, err := filters[0].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase <> ?")
		assert.EqualValues(t, queryExpr.Args, core.WorkflowExecution_RUNNING.String())

		queryExpr, err = filters[1].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase not in (?)")
		assert.EqualValues(t, queryExpr.Args, TerminalWorkflowExecutionPhases)

		queryExpr, err = filters[2].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase in (?)")
		assert.EqualValues(t, queryExpr.Args, SchedulingWorkflowPhases)
	})
	t.Run("terminal", func(t *testing.T) {
		filters, err := GetUpdateExecutionFilters(core.WorkflowExecution_SUCCEEDED)
		assert.NoError(t, err)
		assert.Len(t, filters, 2)

		queryExpr, err := filters[0].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase <> ?")
		assert.EqualValues(t, queryExpr.Args, core.WorkflowExecution_SUCCEEDED.String())

		queryExpr, err = filters[1].GetGormQueryExpr()
		assert.NoError(t, err)
		assert.Equal(t, queryExpr.Query, "phase not in (?)")
		assert.EqualValues(t, queryExpr.Args, TerminalWorkflowExecutionPhases)
	})
}
