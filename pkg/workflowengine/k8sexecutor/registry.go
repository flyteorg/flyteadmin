package k8sexecutor

import (
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/k8sexecutor/impl"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/k8sexecutor/interfaces"
)

var registry = impl.NewRegistry()

func GetRegistry() interfaces.FlyteK8sWorkflowExecutorRegistry {
	return registry
}
