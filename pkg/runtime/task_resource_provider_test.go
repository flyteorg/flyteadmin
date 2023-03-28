package runtime

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewTaskResourceProvider(t *testing.T) {
	tt := &TaskResourceSpec{
		Defaults: interfaces.TaskResourceSet{
			GPU: resource.MustParse("0"),
		},
	}
	fmt.Print(tt.Defaults.GPU.IsZero())
	fmt.Print(tt.Defaults.Storage.IsZero())
}
