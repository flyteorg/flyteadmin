package interfaces

type PropellerExecutorConfig struct {
	ClusterResourceConfig
	NamespaceMappingConfig
}

type PropellerExecutorConfiguration interface {
	GetClusterResourceConfiguration() ClusterResourceConfiguration
	GetNamespaceMappingConfiguration() NamespaceMappingConfiguration
}
