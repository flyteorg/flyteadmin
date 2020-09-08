package common

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

type Entity = string

const (
	Execution           = "e"
	LaunchPlan          = "l"
	NodeExecution       = "ne"
	NodeExecutionEvent  = "nee"
	Task                = "t"
	TaskExecution       = "te"
	Workflow            = "w"
	NamedEntity         = "nen"
	NamedEntityMetadata = "nem"
)

// ResourceTypeToEntity maps a resource type to an entity suitable for use with Database filters
var ResourceTypeToEntity = map[core.ResourceType]Entity{
	core.ResourceType_LAUNCH_PLAN: LaunchPlan,
	core.ResourceType_TASK:        Task,
	core.ResourceType_WORKFLOW:    Workflow,
}
