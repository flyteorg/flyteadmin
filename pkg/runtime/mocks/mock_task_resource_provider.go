package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type MockTaskResourceConfiguration struct {
	Defaults interfaces.TaskResourceSet
	Limits   interfaces.TaskResourceSet
}

func (c *MockTaskResourceConfiguration) GetAsAttribute() admin.TaskResourceAttributes {
	return admin.TaskResourceAttributes{
		Defaults: &admin.TaskResourceSpec{
			Cpu:              c.Defaults.CPU.String(),
			Gpu:              c.Defaults.GPU.String(),
			Memory:           c.Defaults.Memory.String(),
			Storage:          c.Defaults.Storage.String(),
			EphemeralStorage: c.Defaults.EphemeralStorage.String(),
		},
		Limits: &admin.TaskResourceSpec{
			Cpu:              c.Limits.CPU.String(),
			Gpu:              c.Limits.GPU.String(),
			Memory:           c.Limits.Memory.String(),
			Storage:          c.Limits.Storage.String(),
			EphemeralStorage: c.Limits.EphemeralStorage.String(),
		},
	}
}

func (c *MockTaskResourceConfiguration) GetDefaults() interfaces.TaskResourceSet {
	return c.Defaults
}
func (c *MockTaskResourceConfiguration) GetLimits() interfaces.TaskResourceSet {
	return c.Limits
}

func NewMockTaskResourceConfiguration(defaults, limits interfaces.TaskResourceSet) interfaces.TaskResourceConfiguration {
	return &MockTaskResourceConfiguration{
		Defaults: defaults,
		Limits:   limits,
	}
}
