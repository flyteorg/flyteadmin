package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flytestdlib/config"
)

const clusterPoolsKey = "cluster_pools"

var clusterPoolsConfig = config.MustRegisterSection(clusterPoolsKey, &interfaces.ClusterPoolAssignmentConfig{
	ClusterPoolAssignments: make(interfaces.ClusterPoolAssignments),
})

// Implementation of an interfaces.QueueConfiguration
type ClusterPoolAssignmentConfigurationProvider struct{}

func (p *ClusterPoolAssignmentConfigurationProvider) GetClusterPoolAssignments() interfaces.ClusterPoolAssignments {
	return clusterPoolsConfig.GetConfig().(*interfaces.ClusterPoolAssignmentConfig).ClusterPoolAssignments
}

func NewClusterPoolAssignmentConfigurationProvider() interfaces.ClusterPoolAssignmentConfiguration {
	return &ClusterPoolAssignmentConfigurationProvider{}
}
