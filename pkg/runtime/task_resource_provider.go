package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/config"
)

const taskResourceKey = "task_resources"

var taskResourceConfig = config.MustRegisterSection(taskResourceKey, &TaskResourceSpec{
	Defaults:      interfaces.TaskResourceSet{},
	DefaultLimits: interfaces.TaskResourceSet{},
	Limits:        interfaces.TaskResourceSet{},
})

type TaskResourceSpec struct {
	Defaults      interfaces.TaskResourceSet `json:"defaults"`
	DefaultLimits interfaces.TaskResourceSet `json:"default_limits"`
	Limits        interfaces.TaskResourceSet `json:"limits"`
}

// TaskResourceProvider Implementation of an interfaces.TaskResourceConfiguration
type TaskResourceProvider struct{}

func (p *TaskResourceProvider) GetDefaults() interfaces.TaskResourceSet {
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).Defaults
}

func (p *TaskResourceProvider) GetDefaultLimits() interfaces.TaskResourceSet {
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).DefaultLimits
}

func (p *TaskResourceProvider) GetLimits() interfaces.TaskResourceSet {
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).Limits
}

// ConstructTaskResourceSpec takes the configuration struct and turns it into the protobuf struct
func (p *TaskResourceProvider) ConstructTaskResourceSpec(a interfaces.TaskResourceSet) admin.TaskResourceSpec {
	res := admin.TaskResourceSpec{}
	if a.CPU != nil {
		res.Cpu = a.CPU.String()
	}
	if a.GPU != nil {
		res.Gpu = a.GPU.String()
	}
	if a.Memory != nil {
		res.Memory = a.Memory.String()
	}
	if a.EphemeralStorage != nil {
		res.EphemeralStorage = a.EphemeralStorage.String()
	}
	return res
}

func (p *TaskResourceProvider) GetAsAttribute() admin.TaskResourceAttributes {
	defaults := p.ConstructTaskResourceSpec(p.GetDefaults())
	defaultLimits := p.ConstructTaskResourceSpec(p.GetDefaultLimits())
	limits := p.ConstructTaskResourceSpec(p.GetLimits())

	return admin.TaskResourceAttributes{
		Defaults:      &defaults,
		DefaultLimits: &defaultLimits,
		Limits:        &limits,
	}
}

func NewTaskResourceProvider() interfaces.TaskResourceConfiguration {
	return &TaskResourceProvider{}
}
