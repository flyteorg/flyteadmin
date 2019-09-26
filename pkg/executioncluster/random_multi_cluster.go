package executioncluster

import (
	"fmt"

	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/util/rand"
)

type RandomMultiCluster struct {
	executionTargetMap map[string]ExecutionTarget
}

func (s RandomMultiCluster) GetAllValidTargets() []ExecutionTarget {
	v := make([]ExecutionTarget, 0, len(s.executionTargetMap))
	for _, value := range s.executionTargetMap {
		v = append(v, value)
	}
	return v
}

func (s RandomMultiCluster) GetTarget(spec *ExecutionTargetSpec) (*ExecutionTarget, error) {
	if spec != nil {
		if val, ok := s.executionTargetMap[spec.TargetID]; ok {
			return &val, nil
		}
		return nil, fmt.Errorf("invalid cluster target %s", spec.TargetID)
	}
	len := len(s.executionTargetMap)
	targetIdx := rand.Intn(len)
	index := 0
	for _, val := range s.executionTargetMap {
		if index == targetIdx {
			return &val, nil
		}
		index++
	}
	return nil, nil
}

func NewRandomMultiExecutionCluster(scope promutils.Scope, clusterConfig runtime.ClusterConfiguration) (ClusterInterface, error) {
	executionTargetMap, err := GetExecutionTargetMap(scope, clusterConfig)
	if err != nil {
		return nil, err
	}
	return &SingleCluster{
		executionTargetMap: executionTargetMap,
	}, nil
}
