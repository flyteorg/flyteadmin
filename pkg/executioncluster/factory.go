package executioncluster

import (
	"errors"
	"fmt"
	"github.com/lyft/flytestdlib/promutils"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
)

type ClusterInterface interface {
	GetTarget(*ExecutionTargetSpec) (*ExecutionTarget, error)
	GetAllValidTargets() []ExecutionTarget
}

func GetExecutionTargetMap(scope promutils.Scope, clusterConfig runtime.ClusterConfiguration) (map[string]ExecutionTarget, error) {
	executionTargetMap := make(map[string]ExecutionTarget)
	for _, cluster := range clusterConfig.GetClusterConfigs() {
		if _, ok := executionTargetMap[cluster.Name]; ok {
			return nil, errors.New(fmt.Sprintf("duplicate clusters for name %s", cluster.Name))
		}
		executionTarget, err := NewExecutionTarget(scope, cluster)
		if err != nil {
			return nil, err
		}
		executionTargetMap[cluster.Name] = *executionTarget
	}
	return executionTargetMap, nil
}

func GetExecutionCluster(scope promutils.Scope, kubeConfig, master string, clusterConfig runtime.ClusterConfiguration) ClusterInterface {
	switch clusterConfig.GetClusterMode() {
	case runtime.ClusterModeSingle:
		cluster, err := NewSingleExecutionCluster(scope, clusterConfig)
		if err != nil {
			panic(err)
		}
		return cluster
	case runtime.ClusterModeMulti:
		cluster, err := NewRandomMultiExecutionCluster(scope, clusterConfig)
		if err != nil {
			panic(err)
		}
		return cluster
	default:
		cluster, err := NewInCluster(scope, kubeConfig, master)
		if err != nil {
			panic(err)
		}
		return cluster
	}
}
