package runtime

import "flyteadmin/pkg/runtime/interfaces"

type PropellerExecutorConfigurationProvider struct{}

func (p *PropellerExecutorConfigurationProvider) GetClusterResourceConfiguration() ClusterResourceConfigurationProvider {
	return ClusterConfigurationProvider
}

func (p *PropellerExecutorConfigurationProvider) GetNamespaceMappingConfiguration() NamespaceMappingConfigurationProvider {
	return NamespaceMappingConfigurationProvider
}

func NewPropellerExecutorConfigurationProvider() interfaces.PropellerExecutorConfiguration {
	return &PropellerExecutorConfigurationProvider{}
}
