package impl

import (
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/transformers/k8s"
)

type builderMetrics struct {
	Scope                promutils.Scope
	WorkflowBuildSuccess prometheus.Counter
	WorkflowBuildFailure prometheus.Counter
	InvalidExecutionID   prometheus.Counter
}

type flyteWorkflowBuilder struct {
	metrics builderMetrics
}

func (b *flyteWorkflowBuilder) Build(
	wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
	namespace string) (*v1alpha1.FlyteWorkflow, error) {
	return k8s.BuildFlyteWorkflow(wfClosure, inputs, executionID, namespace)
}

func newBuilderMetrics(scope promutils.Scope) builderMetrics {
	return builderMetrics{
		Scope: scope,
		WorkflowBuildSuccess: scope.MustNewCounter("build_success",
			"count of workflows built by propeller without error"),
		WorkflowBuildFailure: scope.MustNewCounter("build_failure",
			"count of workflows built by propeller with errors"),
	}
}

func NewFlyteWorkflowBuilder(scope promutils.Scope) interfaces.FlyteWorkflowBuilder {
	return &flyteWorkflowBuilder{
		metrics: newBuilderMetrics(scope),
	}
}
