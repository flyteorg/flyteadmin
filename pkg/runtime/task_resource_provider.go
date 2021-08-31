package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"k8s.io/apimachinery/pkg/api/resource"
)

const taskResourceKey = "task_resources"

var taskResourceConfig = config.MustRegisterSection(taskResourceKey, &TaskResourceSpec{})

type TaskResourceSpec struct {
	Defaults         interfaces.DeprecatedTaskResourceSet `json:"defaults"`
	DefaultResources interfaces.TaskResourceSet           `json:"defaultResources"`
	Limits           interfaces.DeprecatedTaskResourceSet `json:"limits"`
	LimitResources   interfaces.TaskResourceSet           `json:"limitResources"`
}

// Implementation of an interfaces.TaskResourceConfiguration
type TaskResourceProvider struct{}

func isDeprecatedTaskResourceSet(set interfaces.DeprecatedTaskResourceSet) bool {
	return len(set.CPU) > 0 || len(set.Memory) > 0 || len(set.Storage) > 0 || len(set.EphemeralStorage) > 0 || len(set.GPU) > 0
}

func fromDeprecatedTaskResourceSet(set interfaces.DeprecatedTaskResourceSet) interfaces.TaskResourceSet {
	result := interfaces.TaskResourceSet{}
	if len(set.CPU) > 0 {
		result.CPU = resource.MustParse(set.CPU)
	}
	if len(set.Memory) > 0 {
		result.Memory = resource.MustParse(set.Memory)
	}
	if len(set.Storage) > 0 {
		result.Storage = resource.MustParse(set.Storage)
	}
	if len(set.EphemeralStorage) > 0 {
		result.EphemeralStorage = resource.MustParse(set.EphemeralStorage)
	}
	if len(set.GPU) > 0 {
		result.GPU = resource.MustParse(set.GPU)
	}
	return result
}

func (p *TaskResourceProvider) GetDefaults() interfaces.TaskResourceSet {
	if isDeprecatedTaskResourceSet(taskResourceConfig.GetConfig().(*TaskResourceSpec).Defaults) {
		return fromDeprecatedTaskResourceSet(taskResourceConfig.GetConfig().(*TaskResourceSpec).Defaults)
	}
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).DefaultResources
}

func (p *TaskResourceProvider) GetLimits() interfaces.TaskResourceSet {
	if isDeprecatedTaskResourceSet(taskResourceConfig.GetConfig().(*TaskResourceSpec).Limits) {
		return fromDeprecatedTaskResourceSet(taskResourceConfig.GetConfig().(*TaskResourceSpec).Limits)
	}
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).LimitResources
}

func NewTaskResourceProvider() interfaces.TaskResourceConfiguration {
	return &TaskResourceProvider{}
}
