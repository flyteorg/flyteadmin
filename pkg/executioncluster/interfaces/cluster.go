package interfaces

import (
	"github.com/lyft/flyteadmin/pkg/executioncluster"
)

type ClusterInterface interface {
	GetTarget(*executioncluster.ExecutionTargetSpec) (*executioncluster.ExecutionTarget, error)
	GetAllValidTargets() []executioncluster.ExecutionTarget
}
