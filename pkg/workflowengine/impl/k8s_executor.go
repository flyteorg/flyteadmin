package impl

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	execClusterInterfaces "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var deletePropagationBackground = v1.DeletePropagationBackground

const defaultIdentifier = "DefaultK8sExecutor"

// K8sWorkflowExecutor directly creates and delete Flyte workflow execution CRD objects using the configured execution
// cluster interface.
type K8sWorkflowExecutor struct {
	executionCluster execClusterInterfaces.ClusterInterface
	workflowBuilder  interfaces.FlyteWorkflowBuilder
}

func (d K8sWorkflowExecutor) ID() string {
	return defaultIdentifier
}

func (d K8sWorkflowExecutor) Execute(ctx context.Context, data interfaces.ExecutionData) (interfaces.ExecutionResponse, error) {
	// TODO: Reduce CRD size and use offloaded input URI to blob store instead.
	flyteWf, err := d.workflowBuilder.Build(data.WorkflowClosure.Primary.Template, request.Inputs, data.ExecutionID, data.Namespace)
	if err != nil {
		m.systemMetrics.WorkflowBuildFailures.Inc()
		logger.Infof(ctx, "failed to build the workflow [%+v] %v",
			workflow.Closure.CompiledWorkflow.Primary.Template.Id, err)
		return nil, nil, err
	}

	executionTargetSpec := executioncluster.ExecutionTargetSpec{
		Project:     data.ExecutionID.Project,
		Domain:      data.ExecutionID.Domain,
		Workflow:    data.ReferenceWorkflowName,
		LaunchPlan:  data.ReferenceWorkflowName,
		ExecutionID: data.ExecutionID.Name,
	}
	targetCluster, err := d.executionCluster.GetTarget(ctx, &executionTargetSpec)
	if err != nil {
		return interfaces.ExecutionResponse{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
	}
	_, err = targetCluster.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(data.Namespace).Create(ctx, flyteWf, v1.CreateOptions{})
	if err != nil {
		if !k8_api_err.IsAlreadyExists(err) {
			logger.Debugf(context.TODO(), "Failed to create execution [%+v] in cluster: %s", data.ExecutionID, targetCluster.ID)
			return interfaces.ExecutionResponse{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
		}
	}
	return interfaces.ExecutionResponse{
		Cluster: targetCluster.ID,
	}, nil
}

func (d K8sWorkflowExecutor) Abort(ctx context.Context, data interfaces.AbortData) error {
	target, err := d.executionCluster.GetTarget(ctx, &executioncluster.ExecutionTargetSpec{
		TargetID: data.Cluster,
	})
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, err.Error())
	}
	err = target.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(data.Namespace).Delete(ctx, data.ExecutionID.GetName(), v1.DeleteOptions{
		PropagationPolicy: &deletePropagationBackground,
	})
	// An IsNotFound error indicates the resource is already deleted.
	if err != nil && !k8_api_err.IsNotFound(err) {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to terminate execution: %v with err %v", data.ExecutionID, err)
	}
	return nil
}

func NewK8sWorkflowExecutor(executionCluster execClusterInterfaces.ClusterInterface,
	workflowBuilder interfaces.FlyteWorkflowBuilder) *K8sWorkflowExecutor {

	return &K8sWorkflowExecutor{
		executionCluster: executionCluster,
		workflowBuilder:  workflowBuilder,
	}
}
