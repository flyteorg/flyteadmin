package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

//go:generate mockery -name=K8sWorkflowExecutor -output=../mocks/ -case=underscore

type ExecutionData struct {
	// Execution namespace.
	Namespace string
	// Execution identifier.
	ExecutionID *core.WorkflowExecutionIdentifier
	// Underlying workflow name for the execution.
	ReferenceWorkflowName string
	// Launch plan name used to trigger the execution.
	ReferenceLaunchPlanName string
	// Compiled workflow closure used to build the flyte workflow
	WorkflowClosure *core.CompiledWorkflowClosure
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

// K8sWorkflowExecutor is a client interface used to create and delete Flyte workflow CRD objects.
type K8sWorkflowExecutor interface {
	ID() string
	Execute(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, data ExecutionData) (ExecutionResponse, error)
	Abort(ctx context.Context, data AbortData) error
}
