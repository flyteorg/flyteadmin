package executions

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

const phaseField = "phase"

func newNotTerminalFilter() (common.InlineFilter, error) {
	return common.NewRepeatedValueFilter(common.Execution, common.ValueNotIn, "phase", []string{
		core.WorkflowExecution_ABORTED.String(),
		core.WorkflowExecution_FAILED.String(),
		core.WorkflowExecution_TIMED_OUT.String(),
		core.WorkflowExecution_SUCCEEDED.String(),
	})
}

func newSchedulingFilter() (common.InlineFilter, error) {
	return common.NewRepeatedValueFilter(common.Execution, common.ValueIn, "phase", []string{
		core.WorkflowExecution_UNDEFINED.String(),
		core.WorkflowExecution_QUEUED.String(),
	})
}

func GetUpdateExecutionFilters(eventPhase core.WorkflowExecution_Phase) (filters []common.InlineFilter, err error) {
	// With the exeception of queued events, it's never acceptable to move into the same phase.
	if eventPhase != core.WorkflowExecution_QUEUED {
		notAlreadySamePhaseFilter, err := common.NewSingleValueFilter(common.Execution, common.NotEqual, phaseField, eventPhase.String())
		if err != nil {
			return nil, err
		}
		filters = append(filters, notAlreadySamePhaseFilter)
	}

	// It's never acceptable to move from a terminal state onto another.
	notAlreadyTerminal, err := newNotTerminalFilter()
	if err != nil {
		return nil, err
	}
	filters = append(filters, notAlreadyTerminal)

	if eventPhase == core.WorkflowExecution_QUEUED || eventPhase == core.WorkflowExecution_RUNNING {
		schedulingFilter, err := newSchedulingFilter()
		if err != nil {
			return nil, err
		}
		filters = append(filters, schedulingFilter)
	}
	return filters, nil
}
