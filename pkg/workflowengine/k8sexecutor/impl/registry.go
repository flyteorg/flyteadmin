package impl

import (
	"context"
	"sync"

	"github.com/flyteorg/flyteadmin/pkg/workflowengine/k8sexecutor/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
)

type flyteK8sWorkflowExecutorRegistry struct {
	m               sync.Mutex
	executor        interfaces.K8sWorkflowExecutor
	defaultExecutor interfaces.K8sWorkflowExecutor
}

func (r *flyteK8sWorkflowExecutorRegistry) Register(executor interfaces.K8sWorkflowExecutor) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.executor == nil {
		logger.Debugf(context.TODO(), "setting flyte k8s workflow executor [%s]", executor.ID())
	} else {
		logger.Debugf(context.TODO(), "updating flyte k8s workflow executor [%s]", executor.ID())
	}
	r.executor = executor
}

func (r *flyteK8sWorkflowExecutorRegistry) RegisterDefault(executor interfaces.K8sWorkflowExecutor) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.executor == nil {
		logger.Debugf(context.TODO(), "setting default flyte k8s workflow executor [%s]", executor.ID())
	} else {
		logger.Debugf(context.TODO(), "updating default flyte k8s workflow executor [%s]", executor.ID())
	}
	r.defaultExecutor = executor
}

func (r *flyteK8sWorkflowExecutorRegistry) GetExecutor() interfaces.K8sWorkflowExecutor {
	r.m.Lock()
	defer r.m.Unlock()
	if r.executor == nil {
		return r.defaultExecutor
	}
	return r.executor
}

func NewRegistry() interfaces.FlyteK8sWorkflowExecutorRegistry {
	return &flyteK8sWorkflowExecutorRegistry{}
}
