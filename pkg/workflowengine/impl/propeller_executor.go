package impl

import (
	"context"

	interfaces2 "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/transformers/k8s"
	"google.golang.org/grpc/codes"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
)

var deletePropagationBackground = v1.DeletePropagationBackground

type propellerMetrics struct {
	Scope                     promutils.Scope
	WorkflowBuildSuccess      prometheus.Counter
	WorkflowBuildFailure      prometheus.Counter
	InvalidExecutionID        prometheus.Counter
	ExecutionCreationSuccess  prometheus.Counter
	ExecutionCreationFailure  prometheus.Counter
	TerminateExecutionFailure prometheus.Counter
}

type FlytePropeller struct {
	executionCluster interfaces2.ClusterInterface
	builder          interfaces.FlyteWorkflowInterface
	roleNameKey      string
	metrics          propellerMetrics
	config           runtimeInterfaces.NamespaceMappingConfiguration
	eventVersion     v1alpha1.EventVersion
}

type FlyteWorkflowBuilder struct{}

func (b *FlyteWorkflowBuilder) BuildFlyteWorkflow(
	wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
	namespace string) (*v1alpha1.FlyteWorkflow, error) {
	return k8s.BuildFlyteWorkflow(wfClosure, inputs, executionID, namespace)
}

func addMapValues(overrides map[string]string, flyteWfValues map[string]string) map[string]string {
	if flyteWfValues == nil {
		flyteWfValues = map[string]string{}
	}
	if overrides == nil {
		return flyteWfValues
	}
	for label, value := range overrides {
		flyteWfValues[label] = value
	}
	return flyteWfValues
}

func (c *FlytePropeller) addPermissions(launchPlan admin.LaunchPlan, flyteWf *v1alpha1.FlyteWorkflow) {
	// Set role permissions based on launch plan Auth values.
	// The branched-ness of this check is due to the presence numerous deprecated fields
	var role string
	if launchPlan.Spec.GetAuthRole() != nil && len(launchPlan.GetSpec().GetAuthRole().GetAssumableIamRole()) > 0 {
		role = launchPlan.GetSpec().GetAuthRole().GetAssumableIamRole()
	} else if launchPlan.GetSpec().GetAuth() != nil && len(launchPlan.GetSpec().GetAuth().GetAssumableIamRole()) > 0 {
		role = launchPlan.GetSpec().GetAuth().GetAssumableIamRole()
	} else if len(launchPlan.GetSpec().GetRole()) > 0 {
		// Although deprecated, older launch plans may reference the role field instead of the Auth AssumableIamRole.
		role = launchPlan.GetSpec().GetRole()
	}
	if launchPlan.GetSpec().GetAuthRole() != nil && len(launchPlan.GetSpec().GetAuthRole().GetKubernetesServiceAccount()) > 0 {
		flyteWf.ServiceAccountName = launchPlan.GetSpec().GetAuthRole().GetKubernetesServiceAccount()
	} else if launchPlan.GetSpec().GetAuth() != nil && len(launchPlan.GetSpec().GetAuth().GetKubernetesServiceAccount()) > 0 {
		flyteWf.ServiceAccountName = launchPlan.GetSpec().GetAuth().GetKubernetesServiceAccount()
	}
	if len(role) > 0 {
		if flyteWf.Annotations == nil {
			flyteWf.Annotations = map[string]string{}
		}
		flyteWf.Annotations[c.roleNameKey] = role
	}
}

func addExecutionOverrides(taskPluginOverrides []*admin.PluginOverride, flyteWf *v1alpha1.FlyteWorkflow) {
	executionConfig := v1alpha1.ExecutionConfig{
		TaskPluginImpls: make(map[string]v1alpha1.TaskPluginOverride),
	}
	if len(taskPluginOverrides) == 0 {
		return
	}
	for _, override := range taskPluginOverrides {
		executionConfig.TaskPluginImpls[override.TaskType] = v1alpha1.TaskPluginOverride{
			PluginIDs:             override.PluginId,
			MissingPluginBehavior: override.MissingPluginBehavior,
		}

	}
	flyteWf.ExecutionConfig = executionConfig
}

func (c *FlytePropeller) ExecuteWorkflow(ctx context.Context, input interfaces.ExecuteWorkflowInput) (*interfaces.ExecutionInfo, error) {
	if input.ExecutionID == nil {
		c.metrics.InvalidExecutionID.Inc()
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "invalid execution id")
	}
	namespace := common.GetNamespaceName(c.config.GetNamespaceMappingConfig(), input.ExecutionID.GetProject(), input.ExecutionID.GetDomain())
	flyteWf, err := c.builder.BuildFlyteWorkflow(&input.WfClosure, input.Inputs, input.ExecutionID, namespace)
	if err != nil {
		c.metrics.WorkflowBuildFailure.Inc()
		logger.Infof(ctx, "failed to build the workflow [%+v] %v",
			input.WfClosure.Primary.Template.Id, err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to build the workflow [%+v] %v",
			input.WfClosure.Primary.Template.Id, err)
	}
	c.metrics.WorkflowBuildSuccess.Inc()

	// add the executionId so Propeller can send events back that are associated with the ID
	flyteWf.ExecutionID = v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: input.ExecutionID,
	}
	// add the acceptedAt timestamp so propeller can emit latency metrics.
	acceptAtWrapper := v1.NewTime(input.AcceptedAt)
	flyteWf.AcceptedAt = &acceptAtWrapper

	c.addPermissions(input.Reference, flyteWf)

	labels := addMapValues(input.Labels, flyteWf.Labels)
	flyteWf.Labels = labels
	annotations := addMapValues(input.Annotations, flyteWf.Annotations)
	flyteWf.Annotations = annotations
	if flyteWf.WorkflowMeta == nil {
		flyteWf.WorkflowMeta = &v1alpha1.WorkflowMeta{}
	}
	flyteWf.WorkflowMeta.EventVersion = c.eventVersion
	addExecutionOverrides(input.TaskPluginOverrides, flyteWf)

	if input.Reference.Spec.RawOutputDataConfig != nil {
		flyteWf.RawOutputDataConfig = v1alpha1.RawOutputDataConfig{
			RawOutputDataConfig: input.Reference.Spec.RawOutputDataConfig,
		}
	}

	/*
		TODO(katrogan): uncomment once propeller has updated the flyte workflow CRD.
		queueingBudgetSeconds := int64(input.QueueingBudget.Seconds())
		flyteWf.QueuingBudgetSeconds = &queueingBudgetSeconds
	*/

	executionTargetSpec := executioncluster.ExecutionTargetSpec{
		Project:     input.ExecutionID.Project,
		Domain:      input.ExecutionID.Domain,
		Workflow:    input.Reference.Spec.WorkflowId.Name,
		LaunchPlan:  input.Reference.Id.Name,
		ExecutionID: input.ExecutionID.Name,
	}
	targetCluster, err := c.executionCluster.GetTarget(ctx, &executionTargetSpec)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
	}
	_, err = targetCluster.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Create(ctx, flyteWf, v1.CreateOptions{})
	if err != nil {
		if !k8_api_err.IsAlreadyExists(err) {
			logger.Debugf(ctx, "failed to create workflow [%+v] in cluster %s %v",
				input.WfClosure.Primary.Template.Id, targetCluster.ID, err)
			c.metrics.ExecutionCreationFailure.Inc()
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
		}
	}

	logger.Debugf(ctx, "Successfully created workflow execution [%+v]", input.WfClosure.Primary.Template.Id)
	c.metrics.ExecutionCreationSuccess.Inc()

	return &interfaces.ExecutionInfo{
		Cluster: targetCluster.ID,
	}, nil
}

func (c *FlytePropeller) ExecuteTask(ctx context.Context, input interfaces.ExecuteTaskInput) (*interfaces.ExecutionInfo, error) {
	if input.ExecutionID == nil {
		c.metrics.InvalidExecutionID.Inc()
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "invalid execution id")
	}
	namespace := common.GetNamespaceName(c.config.GetNamespaceMappingConfig(), input.ExecutionID.GetProject(), input.ExecutionID.GetDomain())
	flyteWf, err := c.builder.BuildFlyteWorkflow(&input.WfClosure, input.Inputs, input.ExecutionID, namespace)
	if err != nil {
		c.metrics.WorkflowBuildFailure.Inc()
		logger.Infof(ctx, "failed to build the workflow [%+v] %v",
			input.WfClosure.Primary.Template.Id, err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to build the workflow [%+v] %v",
			input.WfClosure.Primary.Template.Id, err)
	}
	c.metrics.WorkflowBuildSuccess.Inc()

	// add the executionId so Propeller can send events back that are associated with the ID
	flyteWf.ExecutionID = v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: input.ExecutionID,
	}
	// add the acceptedAt timestamp so propeller can emit latency metrics.
	acceptAtWrapper := v1.NewTime(input.AcceptedAt)
	flyteWf.AcceptedAt = &acceptAtWrapper

	// Add execution roles from auth if any.
	if input.Auth != nil && len(input.Auth.GetAssumableIamRole()) > 0 {
		role := input.Auth.GetAssumableIamRole()
		if flyteWf.Annotations == nil {
			flyteWf.Annotations = map[string]string{}
		}
		flyteWf.Annotations[c.roleNameKey] = role
	}
	if input.Auth != nil && len(input.Auth.GetKubernetesServiceAccount()) > 0 {
		flyteWf.ServiceAccountName = input.Auth.GetKubernetesServiceAccount()
	}

	labels := addMapValues(input.Labels, flyteWf.Labels)
	flyteWf.Labels = labels
	annotations := addMapValues(input.Annotations, flyteWf.Annotations)
	flyteWf.Annotations = annotations
	addExecutionOverrides(input.TaskPluginOverrides, flyteWf)

	/*
		TODO(katrogan): uncomment once propeller has updated the flyte workflow CRD.
		queueingBudgetSeconds := int64(input.QueueingBudget.Seconds())
		flyteWf.QueuingBudgetSeconds = &queueingBudgetSeconds
	*/

	executionTargetSpec := executioncluster.ExecutionTargetSpec{
		Project:     input.ExecutionID.Project,
		Domain:      input.ExecutionID.Domain,
		Workflow:    input.ReferenceName,
		LaunchPlan:  input.ReferenceName,
		ExecutionID: input.ExecutionID.Name,
	}
	targetCluster, err := c.executionCluster.GetTarget(ctx, &executionTargetSpec)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
	}
	_, err = targetCluster.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Create(ctx, flyteWf, v1.CreateOptions{})
	if err != nil {
		if !k8_api_err.IsAlreadyExists(err) {
			logger.Debugf(ctx, "failed to create workflow [%+v] in cluster %s %v",
				input.WfClosure.Primary.Template.Id, targetCluster.ID, err)
			c.metrics.ExecutionCreationFailure.Inc()
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
		}
	}

	logger.Debugf(ctx, "Successfully created workflow execution [%+v]", input.WfClosure.Primary.Template.Id)
	c.metrics.ExecutionCreationSuccess.Inc()

	return &interfaces.ExecutionInfo{
		Cluster: targetCluster.ID,
	}, nil
}

func (c *FlytePropeller) TerminateWorkflowExecution(
	ctx context.Context, input interfaces.TerminateWorkflowInput) error {
	if input.ExecutionID == nil {
		c.metrics.InvalidExecutionID.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "invalid execution id")
	}
	namespace := common.GetNamespaceName(c.config.GetNamespaceMappingConfig(), input.ExecutionID.GetProject(), input.ExecutionID.GetDomain())
	target, err := c.executionCluster.GetTarget(ctx, &executioncluster.ExecutionTargetSpec{
		TargetID: input.Cluster,
	})
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, err.Error())
	}
	err = target.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Delete(ctx, input.ExecutionID.GetName(), v1.DeleteOptions{
		PropagationPolicy: &deletePropagationBackground,
	})
	// An IsNotFound error indicates the resource is already deleted.
	if err != nil && !k8_api_err.IsNotFound(err) {
		c.metrics.TerminateExecutionFailure.Inc()
		logger.Errorf(ctx, "failed to terminate execution %v", input.ExecutionID)
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to terminate execution: %v with err %v", input.ExecutionID, err)
	}
	logger.Debugf(ctx, "terminated execution: %v in cluster %s", input.ExecutionID, input.Cluster)
	return nil
}

func newPropellerMetrics(scope promutils.Scope) propellerMetrics {
	return propellerMetrics{
		Scope: scope,
		WorkflowBuildSuccess: scope.MustNewCounter("build_success",
			"count of workflows built by propeller without error"),
		WorkflowBuildFailure: scope.MustNewCounter("build_failure",
			"count of workflows built by propeller with errors"),
		InvalidExecutionID: scope.MustNewCounter("invalid_execution_id",
			"count of invalid execution identifiers used when creating a workflow execution"),
		ExecutionCreationSuccess: scope.MustNewCounter("execution_creation_success",
			"count of successfully created workflow executions"),
		ExecutionCreationFailure: scope.MustNewCounter("execution_creation_failure",
			"count of failed workflow executions creations"),
		TerminateExecutionFailure: scope.MustNewCounter("execution_termination_failure",
			"count of failed workflow executions terminations"),
	}
}

func NewFlytePropeller(roleNameKey string, executionCluster interfaces2.ClusterInterface,
	scope promutils.Scope, configuration runtimeInterfaces.NamespaceMappingConfiguration, eventVersion int) interfaces.Executor {

	return &FlytePropeller{
		executionCluster: executionCluster,
		builder:          &FlyteWorkflowBuilder{},
		roleNameKey:      roleNameKey,
		metrics:          newPropellerMetrics(scope),
		config:           configuration,
		eventVersion:     v1alpha1.EventVersion(eventVersion),
	}
}
