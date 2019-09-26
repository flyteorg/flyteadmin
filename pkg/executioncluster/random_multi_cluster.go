package executioncluster

import (
	"errors"
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
	for  _, value := range s.executionTargetMap {
		v = append(v, value)
	}
	return v
}

func (s RandomMultiCluster) GetTarget(spec *ExecutionTargetSpec) (*ExecutionTarget, error) {
	if spec != nil {
		if val, ok := s.executionTargetMap[spec.TargetId]; ok {
			return &val, nil
		}
		return nil, errors.New(fmt.Sprintf("invalid cluster target %s", spec.TargetId))
	}
	len := len(s.executionTargetMap)
	targetIdx := rand.Intn(len)
	index := 0
	for _, val := range s.executionTargetMap {
		if index == targetIdx {
			return &val, nil
		}
		index += 1
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
