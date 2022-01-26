package impl

import (
	"context"
	"testing"

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
	target := executioncluster.ExecutionTarget{}
	cluster := InCluster{
		target: target,
		asTargets: map[string]*executioncluster.ExecutionTarget{
			target.ID: &target,
		},
	}
	targets := cluster.GetValidTargets()
	assert.Equal(t, 1, len(targets))
	assert.Empty(t, targets[target.ID].ID)

	targets = cluster.GetAllTargets()
	assert.Equal(t, 1, len(targets))
	assert.Empty(t, targets[target.ID].ID)
}
