// Shared method implementations.
package util

import (
	"bytes"
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

var defaultStorageOptions = storage.Options{}

func GetExecutionName(request admin.ExecutionCreateRequest) string {
	if request.Name != "" {
		return request.Name
	}
	return common.GetExecutionName(time.Now().UnixNano())
}

func GetTask(ctx context.Context, repo repositories.RepositoryInterface, identifier core.Identifier) (
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

func CreateDataReference(
	ctx context.Context, identifier *core.Identifier, storagePrefix []string, storageClient *storage.DataStore) (
	storage.DataReference, error) {
	nestedSubKeys := []string{
		identifier.Project,
		identifier.Domain,
		identifier.Name,
		identifier.Version,
	}
	nestedKeys := append(storagePrefix, nestedSubKeys...)
	return storageClient.ConstructReference(ctx, storageClient.GetBaseContainerFQN(ctx), nestedKeys...)
}

func CreateDynamicWorkflowDataReference(
	ctx context.Context, identifier *core.Identifier, nodeExecutionID *core.NodeExecutionIdentifier,
	storagePrefix []string, storageClient *storage.DataStore) (
	storage.DataReference, error) {
	nestedSubKeys := []string{
		identifier.Project,
		identifier.Domain,
		identifier.Name,
		identifier.Version,
		nodeExecutionID.ExecutionId.Project,
		nodeExecutionID.ExecutionId.Domain,
		nodeExecutionID.ExecutionId.Name,
		nodeExecutionID.NodeId,
	}
	nestedKeys := append(storagePrefix, nestedSubKeys...)
	return storageClient.ConstructReference(ctx, storageClient.GetBaseContainerFQN(ctx), nestedKeys...)
}

// This function takes a compiled workflow, and uses the storageClient to write the closure to a location specified by the
// storagePrefix and then finally commits a database entry for workflow.
func WriteCompiledWorkflow(ctx context.Context, repo repositories.RepositoryInterface, remoteClosureDataRef storage.DataReference,
	storageClient *storage.DataStore, id *core.Identifier, workflowClosure *admin.WorkflowClosure) (
	*models.Workflow, error) {
	workflowDigest, err := GetWorkflowDigest(ctx, workflowClosure.CompiledWorkflow)
	if err != nil {
		logger.Errorf(ctx, "failed to compute workflow digest with err %v", err)
		return nil, err
	}

	// Assert that a matching workflow doesn't already exist before uploading the workflow closure.
	existingMatchingWorkflow, err := GetWorkflowModel(ctx, repo, *id)
	// Check that no identical or conflicting workflows exist.
	if err == nil {
		// A workflow's structure is uniquely defined by its collection of nodes.
		if bytes.Equal(workflowDigest, existingMatchingWorkflow.Digest) {
			return nil, errors.NewFlyteAdminErrorf(
				codes.AlreadyExists, "identical workflow already exists with id %v", id)
		}
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"workflow with different structure already exists with id %v", id)
	} else if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
		logger.Debugf(ctx, "Failed to get workflow for comparison in CreateWorkflow with ID [%+v] with err %v",
			id, err)
		return nil, err
	}

	err = storageClient.WriteProtobuf(ctx, remoteClosureDataRef, defaultStorageOptions, workflowClosure)

	if err != nil {
		logger.Infof(ctx,
			"failed to write marshaled workflow with id [%+v] to storage %s with err %v and base container: %s",
			id, remoteClosureDataRef.String(), err, storageClient.GetBaseContainerFQN(ctx))
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to write marshaled workflow [%+v] to storage %s with err %v and base container: %s",
			id, remoteClosureDataRef.String(), err, storageClient.GetBaseContainerFQN(ctx))
	}
	// Save the workflow & its reference to the offloaded, compiled workflow in the database.
	var typedInterface *core.TypedInterface
	if workflowClosure.CompiledWorkflow.Primary != nil && workflowClosure.CompiledWorkflow.Primary.Template != nil {
		typedInterface = workflowClosure.CompiledWorkflow.Primary.Template.Interface
	}
	workflowModel, err := transformers.CreateWorkflowModel(id, remoteClosureDataRef.String(), typedInterface, workflowDigest)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform workflow model for workflow [%+v] and remoteClosureIdentifier [%s] with err: %v",
			id, remoteClosureDataRef.String(), err)
		return nil, err
	}
	if err = repo.WorkflowRepo().Create(ctx, workflowModel); err != nil {
		logger.Infof(ctx, "Failed to create workflow model [%+v] with err %v", id, err)
		return nil, err
	}
	return &workflowModel, nil
}

func GetWorkflowModel(
	ctx context.Context, repo repositories.RepositoryInterface, identifier core.Identifier) (models.Workflow, error) {
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
	repo repositories.RepositoryInterface,
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
	ctx context.Context, repo repositories.RepositoryInterface, identifier core.Identifier) (models.LaunchPlan, error) {
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
	ctx context.Context, repo repositories.RepositoryInterface, identifier core.Identifier) (*admin.LaunchPlan, error) {
	launchPlanModel, err := GetLaunchPlanModel(ctx, repo, identifier)
	if err != nil {
		return nil, err
	}
	return transformers.FromLaunchPlanModel(launchPlanModel)
}

func GetNamedEntityModel(
	ctx context.Context, repo repositories.RepositoryInterface, resourceType core.ResourceType, identifier admin.NamedEntityIdentifier) (models.NamedEntity, error) {
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
	ctx context.Context, repo repositories.RepositoryInterface, resourceType core.ResourceType, identifier admin.NamedEntityIdentifier) (*admin.NamedEntity, error) {
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
	ctx context.Context, repo repositories.RepositoryInterface, identifier core.WorkflowExecutionIdentifier) (
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

func GetNodeExecutionModel(ctx context.Context, repo repositories.RepositoryInterface, nodeExecutionIdentifier *core.NodeExecutionIdentifier) (
	*models.NodeExecution, error) {
	nodeExecutionModel, err := repo.NodeExecutionRepo().Get(ctx, repoInterfaces.NodeExecutionResource{
		NodeExecutionIdentifier: *nodeExecutionIdentifier,
	})

	if err != nil {
		return nil, err
	}
	return &nodeExecutionModel, nil
}

func GetTaskModel(ctx context.Context, repo repositories.RepositoryInterface, taskIdentifier *core.Identifier) (
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
	ctx context.Context, repo repositories.RepositoryInterface, taskExecutionID *core.TaskExecutionIdentifier) (*models.TaskExecution, error) {
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
