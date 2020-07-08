package mocks

import (
	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

type MockQualityOfServiceProvider struct {
	TierExecutionValues map[core.QualityOfService_Tier]core.QualityOfServiceSpec
	DefaultTiers        map[string]core.QualityOfService_Tier
}

func (p MockQualityOfServiceProvider) GetTierExecutionValues() map[core.QualityOfService_Tier]core.QualityOfServiceSpec {
	return p.TierExecutionValues
}

func (p MockQualityOfServiceProvider) GetDefaultTiers() map[string]core.QualityOfService_Tier {
	return p.DefaultTiers
}

func NewMockQualityOfServiceProvider() interfaces.QualityOfServiceConfiguration {
	return &MockQualityOfServiceProvider{
		TierExecutionValues: make(map[core.QualityOfService_Tier]core.QualityOfServiceSpec),
		DefaultTiers:        make(map[string]core.QualityOfService_Tier),
	}
}
