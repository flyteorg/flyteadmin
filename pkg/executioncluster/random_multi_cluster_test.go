package executioncluster

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getRandomeMultiClusterForTest() RandomMultiCluster {
	return RandomMultiCluster{
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

func TestRandomMultiClusterGetTarget(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	target, err := cluster.GetTarget(&ExecutionTargetSpec{TargetId: "cluster-1"})
	assert.Nil(t, err)
	assert.Equal(t, "t1", target.ID)
	target, err = cluster.GetTarget(&ExecutionTargetSpec{TargetId: "cluster-2"})
	assert.Nil(t, err)
	assert.Equal(t, "t2", target.ID)
}

func TestRandomMultiClusterGetRamdomTarget(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	target, err := cluster.GetTarget(nil)
	assert.Nil(t, err)
	assert.NotNil(t, target)
	assert.NotEmpty(t, target.ID)
}

func TestRandomMultiClusterGetRemoteTarget(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	_, err := cluster.GetTarget(&ExecutionTargetSpec{TargetId: "cluster-3"})
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid cluster target cluster-3")
}

func TestRandomMultiClusterGetAllValidTargets(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	targets := cluster.GetAllValidTargets()
	assert.Equal(t, 2, len(targets))
}
