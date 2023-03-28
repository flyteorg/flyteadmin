package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/config"
	"k8s.io/apimachinery/pkg/api/resource"
)

const taskResourceKey = "task_resources"

var taskResourceConfig = config.MustRegisterSection(taskResourceKey, &TaskResourceSpec{
	Defaults: interfaces.TaskResourceSet{
		CPU:    resource.MustParse("2"),
		Memory: resource.MustParse("200Mi"),
		GPU:    resource.MustParse("0"),
	},
	Limits: interfaces.TaskResourceSet{
		CPU:    resource.MustParse("2"),
		Memory: resource.MustParse("1Gi"),
		GPU:    resource.MustParse("1"),
	},
})

type TaskResourceSpec struct {
	Defaults interfaces.TaskResourceSet `json:"defaults"`
	Limits   interfaces.TaskResourceSet `json:"limits"`
}

// TaskResourceProvider Implementation of an interfaces.TaskResourceConfiguration
type TaskResourceProvider struct{}

func (p *TaskResourceProvider) GetDefaults() interfaces.TaskResourceSet {
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).Defaults
}

func (p *TaskResourceProvider) GetLimits() interfaces.TaskResourceSet {
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).Limits
}

func (p *TaskResourceProvider) GetAsAttribute() admin.TaskResourceAttributes {
	defaultCPU := p.GetDefaults().CPU
	defaultCPUStr := defaultCPU.String()
	defaultGPU := p.GetDefaults().GPU
	defaultGPUStr := defaultGPU.String()
	defaultMem := p.GetDefaults().Memory
	defaultMemStr := defaultMem.String()
	defaultEphemeralStorage := p.GetDefaults().EphemeralStorage
	defaultEphemeralStorageStr := defaultEphemeralStorage.String()

	limitCPU := p.GetLimits().CPU
	limitCPUStr := limitCPU.String()
	limitGPU := p.GetLimits().GPU
	limitGPUStr := limitGPU.String()
	limitMem := p.GetLimits().Memory
	limitMemStr := limitMem.String()
	limitEphemeralStorage := p.GetLimits().EphemeralStorage
	limitEphemeralStorageStr := limitEphemeralStorage.String()

	return admin.TaskResourceAttributes{
		Defaults: &admin.TaskResourceSpec{
			Cpu:              defaultCPUStr,
			Gpu:              defaultGPUStr,
			Memory:           defaultMemStr,
			EphemeralStorage: defaultEphemeralStorageStr,
		},
		Limits: &admin.TaskResourceSpec{
			Cpu:              limitCPUStr,
			Gpu:              limitGPUStr,
			Memory:           limitMemStr,
			EphemeralStorage: limitEphemeralStorageStr,
		},
	}
}

func NewTaskResourceProvider() interfaces.TaskResourceConfiguration {
	return &TaskResourceProvider{}
}
