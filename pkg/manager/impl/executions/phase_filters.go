package executions

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
)

const phaseField = "phase"

var TerminalWorkflowExecutionPhases = []string{
	core.WorkflowExecution_ABORTED.String(),
	core.WorkflowExecution_FAILED.String(),
	core.WorkflowExecution_TIMED_OUT.String(),
	core.WorkflowExecution_SUCCEEDED.String(),
}

var SchedulingWorkflowPhases = []string{
	core.WorkflowExecution_UNDEFINED.String(),
	core.WorkflowExecution_QUEUED.String(),
}

var TerminalNodeExecutionPhases = []string{
	core.NodeExecution_SUCCEEDED.String(),
	core.NodeExecution_FAILED.String(),
	core.NodeExecution_TIMED_OUT.String(),
	core.NodeExecution_ABORTED.String(),
	core.NodeExecution_SKIPPED.String(),
	core.NodeExecution_RECOVERED.String(),
}

var TerminalTaskExecutionPhases = []string{
	core.TaskExecution_SUCCEEDED.String(),
	core.TaskExecution_FAILED.String(),
	core.TaskExecution_ABORTED.String(),
}

func newNotTerminalFilter(entity common.Entity) (common.InlineFilter, error) {
	switch entity {
	case common.Execution:
		return common.NewRepeatedValueFilter(common.Execution, common.ValueNotIn, "phase", TerminalWorkflowExecutionPhases)
	case common.NodeExecution:
		return common.NewRepeatedValueFilter(common.Execution, common.ValueNotIn, "phase", TerminalNodeExecutionPhases)
	case common.TaskExecution:
		return common.NewRepeatedValueFilter(common.Execution, common.ValueNotIn, "phase", TerminalTaskExecutionPhases)
	default:
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Unrecognized execution entity [%+v] for non-terminal filters", entity)
	}
}

func newSchedulingFilter() (common.InlineFilter, error) {
	return common.NewRepeatedValueFilter(common.Execution, common.ValueIn, "phase", SchedulingWorkflowPhases)
}

func GetUpdateExecutionFilters(eventPhase core.WorkflowExecution_Phase) (filters []common.InlineFilter, err error) {
	// With the exeception of queued events, it's never acceptable to move into the same phase.
	// With terminal executions we don't need to add the check for the same phase because we always check the existing
	// execution isn't already terminal.
	if eventPhase != core.WorkflowExecution_QUEUED && !common.IsExecutionTerminal(eventPhase) {
		notAlreadySamePhaseFilter, err := common.NewSingleValueFilter(common.Execution, common.NotEqual, phaseField, eventPhase.String())
		if err != nil {
			return nil, err
		}
		filters = append(filters, notAlreadySamePhaseFilter)
	}

	// It's never acceptable to move from a terminal state onto another.
	notAlreadyTerminal, err := newNotTerminalFilter(common.Execution)
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

func GetUpdateNodeExecutionFilters(eventPhase core.NodeExecution_Phase) (filters []common.InlineFilter, err error) {
	// With terminal executions we don't need to add the check for the same phase because we always check the existing
	// execution isn't already terminal.
	if !common.IsNodeExecutionTerminal(eventPhase) {
		notAlreadySamePhaseFilter, err := common.NewSingleValueFilter(common.Execution, common.NotEqual, phaseField, eventPhase.String())
		if err != nil {
			return nil, err
		}
		filters = append(filters, notAlreadySamePhaseFilter)
	}

	notAlreadyTerminal, err := newNotTerminalFilter(common.NodeExecution)
	if err != nil {
		return nil, err
	}
	return append(filters, notAlreadyTerminal), nil
}

func GetUpdateTaskExecutionFilters(eventPhase core.TaskExecution_Phase) (filters []common.InlineFilter, err error) {
	// With terminal executions we don't need to add the check for the same phase because we always check the existing
	// execution isn't already terminal. RUNNING can have multiple phase versions
	if !(common.IsTaskExecutionTerminal(eventPhase) || eventPhase != core.TaskExecution_RUNNING) {
		notAlreadySamePhaseFilter, err := common.NewSingleValueFilter(common.Execution, common.NotEqual, phaseField, eventPhase.String())
		if err != nil {
			return nil, err
		}
		filters = append(filters, notAlreadySamePhaseFilter)
	}

	notAlreadyTerminal, err := newNotTerminalFilter(common.TaskExecution)
	if err != nil {
		return nil, err
	}
	return append(filters, notAlreadyTerminal), nil
}
