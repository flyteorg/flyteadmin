package executioncluster

import (
	"fmt"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
)

type SingleCluster struct {
	currentExecutionTarget ExecutionTarget
	executionTargetMap     map[string]ExecutionTarget
}

func (s SingleCluster) GetTarget(spec *ExecutionTargetSpec) (*ExecutionTarget, error) {
	if spec != nil {
		if val, ok := s.executionTargetMap[spec.TargetID]; ok {
			return &val, nil
		}
		return nil, fmt.Errorf("invalid cluster target %s", spec.TargetID)
	}
	return &s.currentExecutionTarget, nil
}

func (s SingleCluster) GetAllValidTargets() []ExecutionTarget {
	return []ExecutionTarget{
		s.currentExecutionTarget,
	}
}

func NewSingleExecutionCluster(scope promutils.Scope, clusterConfig runtime.ClusterConfiguration) (ClusterInterface, error) {
	var currentClusterConfig *runtime.ClusterConfig
	currentCluster := clusterConfig.GetCurrentCluster()
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if cluster.Name == currentCluster.Name {
			currentClusterConfig = &cluster
		}
	}
	if currentClusterConfig == nil {
		return nil, fmt.Errorf("failed to find current cluster in config")
	}
	currentExecutionTarget, err := NewExecutionTarget(scope, *currentClusterConfig)
	if err != nil {
		return nil, err
	}
	executionTargetMap, err := GetExecutionTargetMap(scope, clusterConfig)
	if err != nil {
		return nil, err
	}
	return &SingleCluster{
		currentExecutionTarget: *currentExecutionTarget,
		executionTargetMap:     executionTargetMap,
	}, nil
}
