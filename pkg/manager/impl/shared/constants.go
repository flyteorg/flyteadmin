// Shared constants for the manager implementation.
package shared

import (
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

// Field names for reference
const (
	Project               = "project"
	Domain                = "domain"
	Name                  = "name"
	ID                    = "id"
	Version               = "version"
	ResourceType          = "resource_type"
	Spec                  = "spec"
	Type                  = "type"
	RuntimeVersion        = "runtime version"
	Metadata              = "metadata"
	TypedInterface        = "typed interface"
	Container             = "container"
	Image                 = "image"
	Limit                 = "limit"
	Offset                = "offset"
	Filters               = "filters"
	ExpectedInputs        = "expected_inputs"
	FixedInputs           = "fixed_inputs"
	DefaultInputs         = "default_inputs"
	Inputs                = "inputs"
	State                 = "state"
	ExecutionID           = "execution_id"
	NodeID                = "node_id"
	NodeExecutionID       = "node_execution_id"
	TaskID                = "task_id"
	OccurredAt            = "occurred_at"
	Event                 = "event"
	ParentTaskExecutionID = "parent_task_execution_id"
	UserInputs            = "user_inputs"
	ProjectDomain         = "project_domain"
)

// Maps a resource type to an entity suitable for use with Database filters
var ResourceTypeToEntity = map[core.ResourceType]common.Entity{
	core.ResourceType_LAUNCH_PLAN: common.LaunchPlan,
	core.ResourceType_TASK:        common.Task,
	core.ResourceType_WORKFLOW:    common.Workflow,
}
