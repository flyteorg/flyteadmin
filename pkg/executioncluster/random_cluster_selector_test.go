package executioncluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getRandomeMultiClusterForTest() RandomClusterSelector {
	return RandomClusterSelector{
		executionTargetMap: map[string]ExecutionTarget{
			"cluster-1": {
				ID:      "t1",
				Enabled: true,
			},
			"cluster-2": {
				ID:      "t2",
				Enabled: true,
			},
			"cluster-disabled": {
				ID: "t4",
			},
		},
		totalEnabledClusterCount: 2,
	}
}

func TestRandomMultiClusterGetTarget(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	target, err := cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-1"})
	assert.Nil(t, err)
	assert.Equal(t, "t1", target.ID)
	assert.True(t, target.Enabled)
	target, err = cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-disabled"})
	assert.Nil(t, err)
	assert.Equal(t, "t4", target.ID)
	assert.False(t, target.Enabled)
}

func TestRandomMultiClusterGetRandomTarget(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	target, err := cluster.GetTarget(nil)
	assert.Nil(t, err)
	assert.NotNil(t, target)
	assert.NotEmpty(t, target.ID)
}

func TestRandomMultiClusterGetRemoteTarget(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	_, err := cluster.GetTarget(&ExecutionTargetSpec{TargetID: "cluster-3"})
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid cluster target cluster-3")
}

func TestRandomMultiClusterGetAllValidTargets(t *testing.T) {
	cluster := getRandomeMultiClusterForTest()
	targets := cluster.GetAllValidTargets()
	assert.Equal(t, 2, len(targets))
}
