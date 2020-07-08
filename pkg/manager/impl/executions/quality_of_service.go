package executions

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

type QualityOfServiceSpec struct {
	QueuingBudget time.Duration
}

type GetQualityOfServiceInput struct {
	Workflow               *admin.Workflow
	LaunchPlan             *admin.LaunchPlan
	ExecutionCreateRequest *admin.ExecutionCreateRequest
}

type QualityOfServiceAllocator interface {
	GetQualityOfService(ctx context.Context, input GetQualityOfServiceInput) (QualityOfServiceSpec, error)
}

type qualityOfServiceAllocator struct {
	config          runtimeInterfaces.Configuration
	resourceManager interfaces.ResourceInterface
}

func (q qualityOfServiceAllocator) GetQualityOfService(ctx context.Context, input GetQualityOfServiceInput) (QualityOfServiceSpec, error) {
	workflowIdentifier := input.Workflow.Id

	var qualityOfServiceTier core.QualityOfService_Tier
	if input.ExecutionCreateRequest.Spec.QualityOfService != nil {
		if input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec() != nil {
			duration, err := ptypes.Duration(input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget)
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in create execution request [%s/%s/%s], failed to parse duration [%v] with: %v",
					input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
					input.ExecutionCreateRequest.Name,
					input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget, err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.ExecutionCreateRequest.Spec.QualityOfService.GetTier()
	} else if input.LaunchPlan.Spec.QualityOfService != nil {
		if input.LaunchPlan.Spec.QualityOfService.GetSpec() != nil {
			duration, err := ptypes.Duration(input.LaunchPlan.Spec.QualityOfService.GetSpec().QueueingBudget)
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in launch plan [%v], failed to parse duration [%v] with: %v",
					input.LaunchPlan.Id,
					input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget, err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.LaunchPlan.Spec.QualityOfService.GetTier()
	} else if input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata != nil &&
		input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService != nil {
		if input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService.GetSpec() != nil {
			duration, err := ptypes.Duration(input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService.
				GetSpec().QueueingBudget)
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in workflow [%v], failed to parse duration [%v] with: %v",
					workflowIdentifier,
					input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget, err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService.GetTier()
	}

	// If nothing in the hierarchy has set the quality of service, see if an override exists in the matchable attributes
	// resource table.
	resource, err := q.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      workflowIdentifier.Project,
		Domain:       workflowIdentifier.Domain,
		Workflow:     workflowIdentifier.Name,
		ResourceType: admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION,
	})
	if err != nil {
		if _, ok := err.(errors.FlyteAdminError); !ok || err.(errors.FlyteAdminError).Code() != codes.NotFound {
			logger.Warningf(ctx,
				"Failed to fetch override values when assigning quality of service values for [%+v] with err: %v",
				workflowIdentifier, err)
		}
	}

	if resource != nil && resource.Attributes != nil && resource.Attributes.GetQualityOfService() != nil &&
		resource.Attributes.GetQualityOfService().GetSpec() != nil {
		// Use custom override value in database rather than from registered entities or the admin application config.
		duration, err := ptypes.Duration(resource.Attributes.GetQualityOfService().GetSpec().QueueingBudget)
		if err != nil {
			return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"Invalid custom quality of service set for [%+v], failed to parse duration [%v] with: %v",
				workflowIdentifier, resource.Attributes.GetQualityOfService().GetSpec().QueueingBudget, err)
		}
		return QualityOfServiceSpec{
			QueuingBudget: duration,
		}, nil
	} else if resource != nil && resource.Attributes != nil && resource.Attributes.GetQualityOfService() != nil &&
		resource.Attributes.GetQualityOfService().GetTier() != core.QualityOfService_UNDEFINED {
		qualityOfServiceTier = resource.Attributes.GetQualityOfService().GetTier()
	}

	if qualityOfServiceTier == core.QualityOfService_UNDEFINED {
		// If we've come all this way and at no layer is an overridable configuration for the quality of service tier
		// set, use the default values from the admin application config.
		var ok bool
		qualityOfServiceTier, ok = q.config.QualityOfServiceConfiguration().GetDefaultTiers()[input.ExecutionCreateRequest.Domain]
		if !ok {
			// No queueing budget to set when no default is specified
			return QualityOfServiceSpec{}, nil
		}
	}
	executionValues, ok := q.config.QualityOfServiceConfiguration().GetTierExecutionValues()[qualityOfServiceTier]
	if !ok {
		// No queueing budget to set when no default is specified
		return QualityOfServiceSpec{}, nil
	}
	// Config values should always be vetted so there's no need to check the error from conversion.
	duration, _ := ptypes.Duration(executionValues.QueueingBudget)

	return QualityOfServiceSpec{
		QueuingBudget: duration,
	}, nil
}

func NewQualityOfServiceAllocator(config runtimeInterfaces.Configuration, resourceManager interfaces.ResourceInterface) QualityOfServiceAllocator {
	return &qualityOfServiceAllocator{
		config:          config,
		resourceManager: resourceManager,
	}
}
