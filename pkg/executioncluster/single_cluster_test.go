package executioncluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getSingleClusterForTest() SingleCluster {
	return SingleCluster{
		currentExecutionTarget: ExecutionTarget{
			ID: "t1",
		},
		executionTargetMap: map[string]ExecutionTarget{
			"cluster-1": {
				ID: "t1",
			},
			"cluster-2": {
				ID: "t2",
			},
		},
	}
}

func TestSingleClusterGetTarget(t *testing.T) {
	cluster := getSingleClusterForTest()
	target, err := cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-1"})
	assert.Nil(t, err)
	assert.Equal(t, "t1", target.ID)
	target, err = cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-2"})
	assert.Nil(t, err)
	assert.Equal(t, "t2", target.ID)
}

func TestSingleClusterGetRamdomTarget(t *testing.T) {
	cluster := getSingleClusterForTest()
	target, err := cluster.GetTarget(nil)
	assert.Nil(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, "t1", target.ID)
}

func TestSingleClusterGetRemoteTarget(t *testing.T) {
	cluster := getSingleClusterForTest()
	_, err := cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-3"})
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid cluster target cluster-3")
}

func TestSingleClusterGetAllValidTargets(t *testing.T) {
	cluster := getSingleClusterForTest()
	targets := cluster.GetAllValidTargets()
	assert.Equal(t, 1, len(targets))
	assert.Equal(t, "t1", targets[0].ID)
}
