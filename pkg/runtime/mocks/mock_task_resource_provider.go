package mocks

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type MockTaskResourceConfiguration struct {
	Defaults      interfaces.TaskResourceSet
	DefaultLimits interfaces.TaskResourceSet
	Limits        interfaces.TaskResourceSet
}

func (c *MockTaskResourceConfiguration) ConstructTaskResourceSpec(a interfaces.TaskResourceSet) admin.TaskResourceSpec {
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

func (c *MockTaskResourceConfiguration) GetAsAttribute() admin.TaskResourceAttributes {
	defaults := c.ConstructTaskResourceSpec(c.GetDefaults())
	defaultLimits := c.ConstructTaskResourceSpec(c.GetDefaultLimits())
	limits := c.ConstructTaskResourceSpec(c.GetLimits())

	return admin.TaskResourceAttributes{
		Defaults:      &defaults,
		DefaultLimits: &defaultLimits,
		Limits:        &limits,
	}
}

func (c *MockTaskResourceConfiguration) GetDefaults() interfaces.TaskResourceSet {
	return c.Defaults
}
func (c *MockTaskResourceConfiguration) GetLimits() interfaces.TaskResourceSet {
	return c.Limits
}
func (c *MockTaskResourceConfiguration) GetDefaultLimits() interfaces.TaskResourceSet {
	return c.DefaultLimits
}

func NewMockTaskResourceConfiguration(defaults, defaultLimits, limits interfaces.TaskResourceSet) interfaces.TaskResourceConfiguration {
	return &MockTaskResourceConfiguration{
		Defaults:      defaults,
		DefaultLimits: defaultLimits,
		Limits:        limits,
	}
}
