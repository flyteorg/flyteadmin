package flytek8s

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


var deletePropagationBackground = v1.DeletePropagationBackground

type defaultWorkflowExecutor struct {
	executionCluster interfaces.ClusterInterface
}

func (d defaultWorkflowExecutor) Execute(ctx context.Context, flyteWf *v1alpha1.FlyteWorkflow, data ExecutionData) (ExecutionResponse, error) {
	executionTargetSpec := executioncluster.ExecutionTargetSpec{
		Project:     data.ExecutionID.Project,
		Domain:      data.ExecutionID.Domain,
		Workflow:    data.ReferenceWorkflowName,
		LaunchPlan:  data.ReferenceWorkflowName,
		ExecutionID: data.ExecutionID.Name,
	}
	targetCluster, err := d.executionCluster.GetTarget(ctx, &executionTargetSpec)
	if err != nil {
		return ExecutionResponse{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
	}
	_, err = targetCluster.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(data.Namespace).Create(ctx, flyteWf, v1.CreateOptions{})
	if err != nil {
		if !k8_api_err.IsAlreadyExists(err) {
			logger.Debugf(context.TODO(), "Failed to create execution [%+v] in cluster: %s", data.ExecutionID, targetCluster.ID)
			return ExecutionResponse{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
		}
	}
	return ExecutionResponse{
		Cluster: targetCluster.ID,
	}, nil
}

func (d defaultWorkflowExecutor) Abort(ctx context.Context, data AbortData) error {
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

func RegisterDefaultWorkflowExecutor(execClusterInterface interfaces.ClusterInterface) {
	executor := defaultWorkflowExecutor{
		executionCluster: execClusterInterface,
	}
	GetRegistry().Register(executor)
}


