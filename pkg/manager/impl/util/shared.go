// Shared method implementations.
package util

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

func GetExecutionName(request admin.ExecutionCreateRequest) string {
	if request.Name != "" {
		return request.Name
	}
	return common.GetExecutionName(time.Now().UnixNano())
}

func GetTask(ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (
	*admin.Task, error) {
	taskModel, err := GetTaskModel(ctx, repo, &identifier)
	if err != nil {
		return nil, err
	}
	task, err := transformers.FromTaskModel(*taskModel)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform task model for identifier [%+v] with err: %v", identifier, err)
		return nil, err
	}
	return &task, nil
}

func GetWorkflowModel(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (models.Workflow, error) {
	workflowModel, err := (repo).WorkflowRepo().Get(ctx, repoInterfaces.Identifier{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
		Version: identifier.Version,
	})
	if err != nil {
		return models.Workflow{}, err
	}
	return workflowModel, nil
}

func FetchAndGetWorkflowClosure(ctx context.Context,
	store *storage.DataStore,
	remoteLocationIdentifier string) (*admin.WorkflowClosure, error) {
	closure := &admin.WorkflowClosure{}

	err := store.ReadProtobuf(ctx, storage.DataReference(remoteLocationIdentifier), closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"Unable to read WorkflowClosure from location %s : %v", remoteLocationIdentifier, err)
	}
	return closure, nil
}

func GetWorkflow(
	ctx context.Context,
	repo repoInterfaces.Repository,
	store *storage.DataStore,
	identifier core.Identifier) (*admin.Workflow, error) {
	workflowModel, err := GetWorkflowModel(ctx, repo, identifier)
	if err != nil {
		return nil, err
	}
	workflow, err := transformers.FromWorkflowModel(workflowModel)
	if err != nil {
		return nil, err
	}
	closure, err := FetchAndGetWorkflowClosure(ctx, store, workflowModel.RemoteClosureIdentifier)
	if err != nil {
		return nil, err
	}
	closure.CreatedAt = workflow.Closure.CreatedAt
	workflow.Closure = closure
	return &workflow, nil
}

func GetLaunchPlanModel(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (models.LaunchPlan, error) {
	launchPlanModel, err := (repo).LaunchPlanRepo().Get(ctx, repoInterfaces.Identifier{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
		Version: identifier.Version,
	})
	if err != nil {
		return models.LaunchPlan{}, err
	}
	return launchPlanModel, nil
}

func GetLaunchPlan(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (*admin.LaunchPlan, error) {
	launchPlanModel, err := GetLaunchPlanModel(ctx, repo, identifier)
	if err != nil {
		return nil, err
	}
	return transformers.FromLaunchPlanModel(launchPlanModel)
}

func GetNamedEntityModel(
	ctx context.Context, repo repoInterfaces.Repository, resourceType core.ResourceType, identifier admin.NamedEntityIdentifier) (models.NamedEntity, error) {
	metadataModel, err := (repo).NamedEntityRepo().Get(ctx, repoInterfaces.GetNamedEntityInput{
		ResourceType: resourceType,
		Project:      identifier.Project,
		Domain:       identifier.Domain,
		Name:         identifier.Name,
	})
	if err != nil {
		return models.NamedEntity{}, err
	}
	return metadataModel, nil
}

func GetNamedEntity(
	ctx context.Context, repo repoInterfaces.Repository, resourceType core.ResourceType, identifier admin.NamedEntityIdentifier) (*admin.NamedEntity, error) {
	metadataModel, err := GetNamedEntityModel(ctx, repo, resourceType, identifier)
	if err != nil {
		return nil, err
	}
	metadata := transformers.FromNamedEntityModel(metadataModel)
	return &metadata, nil
}

// Returns the set of filters necessary to query launch plan models to find the active version of a launch plan
func GetActiveLaunchPlanVersionFilters(project, domain, name string) ([]common.InlineFilter, error) {
	projectFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Project, project)
	if err != nil {
		return nil, err
	}
	domainFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Domain, domain)
	if err != nil {
		return nil, err
	}
	nameFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Name, name)
	if err != nil {
		return nil, err
	}
	activeFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.State, int32(admin.LaunchPlanState_ACTIVE))
	if err != nil {
		return nil, err
	}
	return []common.InlineFilter{projectFilter, domainFilter, nameFilter, activeFilter}, nil
}

// Returns the set of filters necessary to query launch plan models to find the active version of a launch plan
func ListActiveLaunchPlanVersionsFilters(project, domain string) ([]common.InlineFilter, error) {
	projectFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Project, project)
	if err != nil {
		return nil, err
	}
	domainFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Domain, domain)
	if err != nil {
		return nil, err
	}
	activeFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.State, int32(admin.LaunchPlanState_ACTIVE))
	if err != nil {
		return nil, err
	}
	return []common.InlineFilter{projectFilter, domainFilter, activeFilter}, nil
}

func GetExecutionModel(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.WorkflowExecutionIdentifier) (
	*models.Execution, error) {
	executionModel, err := repo.ExecutionRepo().Get(ctx, repoInterfaces.Identifier{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
	})
	if err != nil {
		return nil, err
	}
	return &executionModel, nil
}

func GetNodeExecutionModel(ctx context.Context, repo repoInterfaces.Repository, nodeExecutionIdentifier *core.NodeExecutionIdentifier) (
	*models.NodeExecution, error) {
	nodeExecutionModel, err := repo.NodeExecutionRepo().Get(ctx, repoInterfaces.NodeExecutionResource{
		NodeExecutionIdentifier: *nodeExecutionIdentifier,
	})

	if err != nil {
		return nil, err
	}
	return &nodeExecutionModel, nil
}

func GetTaskModel(ctx context.Context, repo repoInterfaces.Repository, taskIdentifier *core.Identifier) (
	*models.Task, error) {

	taskModel, err := repo.TaskRepo().Get(ctx, repoInterfaces.Identifier{
		Project: taskIdentifier.Project,
		Domain:  taskIdentifier.Domain,
		Name:    taskIdentifier.Name,
		Version: taskIdentifier.Version,
	})

	if err != nil {
		return nil, err
	}
	return &taskModel, nil
}

func GetTaskExecutionModel(
	ctx context.Context, repo repoInterfaces.Repository, taskExecutionID *core.TaskExecutionIdentifier) (*models.TaskExecution, error) {
	if err := validation.ValidateTaskExecutionIdentifier(taskExecutionID); err != nil {
		logger.Debugf(ctx, "can't get task execution with invalid identifier [%v]: %v", taskExecutionID, err)
		return nil, err
	}

	taskExecutionModel, err := repo.TaskExecutionRepo().Get(ctx, repoInterfaces.GetTaskExecutionInput{
		TaskExecutionID: *taskExecutionID,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to get task execution with id [%+v] with err %v", taskExecutionID, err)
		return nil, err
	}
	return &taskExecutionModel, nil
}

// GetMatchableResource gets matchable resource for resourceType and project - domain combination.
// Returns an error if such a resource is not found.
func GetMatchableResource(ctx context.Context, resourceManager interfaces.ResourceInterface, resourceType admin.MatchableResource,
	project, domain string) (*interfaces.ResourceResponse, error) {
	matchableResource, err := resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: resourceType,
	})
	if err != nil {
		if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
			logger.Errorf(ctx, "Failed to get %v overrides in %s project %s domain with error: %v", resourceType,
				project, domain, err)
			return nil, err
		}
	}
	return matchableResource, nil
}
