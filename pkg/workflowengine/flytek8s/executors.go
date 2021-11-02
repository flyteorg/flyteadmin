package flytek8s

import (
	"context"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/logger"
	"sync"
)

type ExecutionData struct {
	// Execution namespace.
	Namespace string
	// Execution identifier.
	ExecutionID *core.WorkflowExecutionIdentifier
	// Underlying workflow name for the execution.
	ReferenceWorkflowName string
	// Launch plan name used to trigger the execution.
	ReferenceLaunchPlanName string
}

type ExecutionResponse struct {
	// Cluster identifier where the execution was created
	Cluster string
}

type AbortData struct {
	// Execution namespace.
	Namespace string
	// Execution identifier.
	ExecutionID *core.WorkflowExecutionIdentifier
	// Cluster identifier where the execution was created
	Cluster string
}

//go:generate mockery -name=FlyteK8sWorkflowExecutor -output=mocks/ -case=underscore

// FlyteK8sWorkflowExecutor is a client interface used to create and delete Flyte workflow CRD objects.
type FlyteK8sWorkflowExecutor interface {
	Execute(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, data ExecutionData) (ExecutionResponse, error)
	Abort(ctx context.Context, data AbortData) error
}

// FlyteK8sWorkflowExecutorRegistry is a singleton source of which FlyteK8sWorkflowExecutor implementation to use for
// creating and deleting Flyte workflow CRD objects.
type FlyteK8sWorkflowExecutorRegistry interface {
	Register(executor FlyteK8sWorkflowExecutor)
	GetExecutor() FlyteK8sWorkflowExecutor
}

type flyteK8sWorkflowExecutorRegistry struct {
	m        sync.Mutex
	executor FlyteK8sWorkflowExecutor
}

func (r *flyteK8sWorkflowExecutorRegistry) Register(executor FlyteK8sWorkflowExecutor) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.executor == nil {
		logger.Debugf(context.TODO(), "setting flyte k8s workflow executor")
	} else {
		logger.Debugf(context.TODO(), "updating flyte k8s workflow executor")
	}
	r.executor = executor
}

func (r *flyteK8sWorkflowExecutorRegistry) GetExecutor() FlyteK8sWorkflowExecutor {
	r.m.Lock()
	defer r.m.Unlock()
	return r.executor
}

var registry = &flyteK8sWorkflowExecutorRegistry{}

func GetRegistry() FlyteK8sWorkflowExecutorRegistry {
	return registry
}