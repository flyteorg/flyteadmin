package executioncluster

import (
	"errors"
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
		if val, ok := s.executionTargetMap[spec.TargetId]; ok {
			return &val, nil
		}
		return nil, errors.New(fmt.Sprintf("invalid cluster target %s", spec.TargetId))
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
		return nil, errors.New("failed to find current cluster in config")
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
