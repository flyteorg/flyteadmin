package impl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster/mocks"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"

	"github.com/stretchr/testify/assert"
)

func TestInClusterGetTarget(t *testing.T) {
	cluster := InCluster{
		target: executioncluster.ExecutionTarget{
			ID: "t1",
		},
	}
	target, err := cluster.GetTarget(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, "t1", target.ID)
}

func TestInClusterGetRemoteTarget(t *testing.T) {
	cluster := InCluster{
		target: executioncluster.ExecutionTarget{},
	}
	_, err := cluster.GetTarget(context.Background(), &executioncluster.ExecutionTargetSpec{TargetID: "t1"})
	assert.EqualError(t, err, "remote target t1 is not supported")
}

func TestInClusterGetAllValidTargets(t *testing.T) {
	target := executioncluster.ExecutionTarget{
		ID: "t1",
	}
	listTargetsProvider := mocks.ListTargetsInterface{}
	listTargetsProvider.OnGetAllValidTargets().Return(map[string]executioncluster.ExecutionTarget{
		"t1": target,
	})
	cluster := InCluster{
		ListTargetsInterface: &listTargetsProvider,
		target:               target,
	}
	targets := cluster.GetAllValidTargets()
	assert.Equal(t, 1, len(targets))
	assert.Equal(t, "t1", targets["t1"].ID)
}
