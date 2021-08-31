package runtime

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestIsDeprecatedTaskResourceSet(t *testing.T) {
	assert.True(t, isDeprecatedTaskResourceSet(interfaces.DeprecatedTaskResourceSet{
		CPU: "1",
	}))
	assert.False(t, isDeprecatedTaskResourceSet(interfaces.DeprecatedTaskResourceSet{}))
}

func TestFromDeprecatedTaskResourceSet(t *testing.T) {
	taskResourceSet := fromDeprecatedTaskResourceSet(interfaces.DeprecatedTaskResourceSet{
		CPU:              "1",
		Memory:           "100",
		Storage:          "200",
		EphemeralStorage: "300",
		GPU:              "2",
	})
	assert.EqualValues(t, interfaces.TaskResourceSet{
		CPU:              resource.MustParse("1"),
		Memory:           resource.MustParse("100"),
		Storage:          resource.MustParse("200"),
		EphemeralStorage: resource.MustParse("300"),
		GPU:              resource.MustParse("2"),
	}, taskResourceSet)
}
