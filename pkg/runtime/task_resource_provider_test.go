package runtime

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common/testutils"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewTaskResourceProvider(t *testing.T) {
	tt := &TaskResourceSpec{
		Defaults: interfaces.TaskResourceSet{
			GPU: testutils.GetPtr(resource.MustParse("0")),
		},
	}
	assert.True(t, tt.Defaults.GPU.IsZero())
	assert.Nil(t, tt.Defaults.Storage)
}
