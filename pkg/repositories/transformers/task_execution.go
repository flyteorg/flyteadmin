package transformers

import (
	"context"
	"strconv"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/storage"

	"google.golang.org/protobuf/encoding/protojson"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	_struct "github.com/golang/protobuf/ptypes/struct"

	"google.golang.org/grpc/codes"
)

var empty _struct.Struct
var jsonEmpty, _ = protojson.Marshal(&empty)

type CreateTaskExecutionModelInput struct {
	Request               *admin.TaskExecutionEventRequest
	InlineEventDataPolicy interfaces.InlineEventDataPolicy
	StorageClient         *storage.DataStore
}

func addTaskStartedState(request *admin.TaskExecutionEventRequest, taskExecutionModel *models.TaskExecution,
	closure *admin.TaskExecutionClosure) error {
	occurredAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal occurredAt with error: %v", err)
	}
	taskExecutionModel.StartedAt = &occurredAt
	closure.StartedAt = request.Event.OccurredAt
	return nil
}

func addTaskTerminalState(
	ctx context.Context,
	request *admin.TaskExecutionEventRequest,
	taskExecutionModel *models.TaskExecution, closure *admin.TaskExecutionClosure,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	if taskExecutionModel.StartedAt == nil {
		logger.Warning(context.Background(), "task execution is missing StartedAt")
	} else {
		endTime, err := ptypes.Timestamp(request.Event.OccurredAt)
		if err != nil {
			return errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to parse task execution occurredAt timestamp: %v", err)
		}
		closure.StartedAt, err = ptypes.TimestampProto(*taskExecutionModel.StartedAt)
		if err != nil {
			return errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to parse task execution startedAt timestamp: %v", err)
		}
		taskExecutionModel.Duration = endTime.Sub(*taskExecutionModel.StartedAt)
		closure.Duration = ptypes.DurationProto(taskExecutionModel.Duration)
	}

	if request.Event.GetOutputUri() != "" {
		closure.OutputResult = &admin.TaskExecutionClosure_OutputUri{
			OutputUri: request.Event.GetOutputUri(),
		}
	} else if request.Event.GetOutputData() != nil {
		switch inlineEventDataPolicy {
		case interfaces.InlineEventDataPolicyStoreInline:
			closure.OutputResult = &admin.TaskExecutionClosure_OutputData{
				OutputData: request.Event.GetOutputData(),
			}
		default:
			logger.Debugf(ctx, "Offloading outputs per InlineEventDataPolicy")
			uri, err := common.OffloadLiteralMap(ctx, storageClient, request.Event.GetOutputData(),
				request.Event.ParentNodeExecutionId.ExecutionId.Project, request.Event.ParentNodeExecutionId.ExecutionId.Domain,
				request.Event.ParentNodeExecutionId.ExecutionId.Name, request.Event.ParentNodeExecutionId.NodeId,
				request.Event.TaskId.Project, request.Event.TaskId.Domain, request.Event.TaskId.Name, request.Event.TaskId.Version,
				strconv.FormatUint(uint64(request.Event.RetryAttempt), 10), OutputsObjectSuffix)
			if err != nil {
				return err
			}
			closure.OutputResult = &admin.TaskExecutionClosure_OutputUri{
				OutputUri: uri.String(),
			}
		}
	} else if request.Event.GetError() != nil {
		closure.OutputResult = &admin.TaskExecutionClosure_Error{
			Error: request.Event.GetError(),
		}
	}
	return nil
}

func CreateTaskExecutionModel(ctx context.Context, input CreateTaskExecutionModelInput) (*models.TaskExecution, error) {
	taskExecution := &models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: input.Request.Event.TaskId.Project,
				Domain:  input.Request.Event.TaskId.Domain,
				Name:    input.Request.Event.TaskId.Name,
				Version: input.Request.Event.TaskId.Version,
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: input.Request.Event.ParentNodeExecutionId.NodeId,
				ExecutionKey: models.ExecutionKey{
					Project: input.Request.Event.ParentNodeExecutionId.ExecutionId.Project,
					Domain:  input.Request.Event.ParentNodeExecutionId.ExecutionId.Domain,
					Name:    input.Request.Event.ParentNodeExecutionId.ExecutionId.Name,
				},
			},
			RetryAttempt: &input.Request.Event.RetryAttempt,
		},

		Phase:        input.Request.Event.Phase.String(),
		PhaseVersion: input.Request.Event.PhaseVersion,
		InputURI:     input.Request.Event.InputUri,
	}

	closure := &admin.TaskExecutionClosure{
		Phase:      input.Request.Event.Phase,
		UpdatedAt:  input.Request.Event.OccurredAt,
		CreatedAt:  input.Request.Event.OccurredAt,
		Logs:       input.Request.Event.Logs,
		CustomInfo: input.Request.Event.CustomInfo,
		Reason:     input.Request.Event.Reason,
		TaskType:   input.Request.Event.TaskType,
		Metadata:   input.Request.Event.Metadata,
	}

	eventPhase := input.Request.Event.Phase

	// Different tasks may report different phases as their first event.
	// If the first event we receive for this execution is a valid
	// non-terminal phase, mark the execution start time.
	if eventPhase == core.TaskExecution_RUNNING {
		err := addTaskStartedState(input.Request, taskExecution, closure)
		if err != nil {
			return nil, err
		}
	}

	if common.IsTaskExecutionTerminal(input.Request.Event.Phase) {
		err := addTaskTerminalState(ctx, input.Request, taskExecution, closure, input.InlineEventDataPolicy, input.StorageClient)
		if err != nil {
			return nil, err
		}
	}
	marshaledClosure, err := proto.Marshal(closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal task execution closure with error: %v", err)
	}

	taskExecution.Closure = marshaledClosure
	taskExecutionCreatedAt, err := ptypes.Timestamp(input.Request.Event.OccurredAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read event timestamp")
	}
	taskExecution.TaskExecutionCreatedAt = &taskExecutionCreatedAt
	taskExecution.TaskExecutionUpdatedAt = &taskExecutionCreatedAt

	return taskExecution, nil
}

// mergeLogs returns the unique list of logs across an existing list and the latest list sent in a task execution event
// update.
// It returns all the new logs receives + any existing log that hasn't been overwritten by a new log.
// An existing logLink is said to have been overwritten if a new logLink with the same Uri or the same Name has been
// received.
func mergeLogs(existing, latest []*core.TaskLog) []*core.TaskLog {
	if len(latest) == 0 {
		return existing
	}

	if len(existing) == 0 {
		return latest
	}

	latestSetByURI := make(map[string]*core.TaskLog, len(latest))
	latestSetByName := make(map[string]*core.TaskLog, len(latest))
	for _, latestLog := range latest {
		latestSetByURI[latestLog.Uri] = latestLog
		if len(latestLog.Name) > 0 {
			latestSetByName[latestLog.Name] = latestLog
		}
	}

	// Copy over the latest logs since names will change for existing logs as a task transitions across phases.
	logs := latest
	for _, existingLog := range existing {
		if _, ok := latestSetByURI[existingLog.Uri]; !ok {
			if _, ok = latestSetByName[existingLog.Name]; !ok {
				// We haven't seen this log before: add it to the output result list.
				logs = append(logs, existingLog)
			}
		}
	}

	return logs
}

func mergeCustom(existing, latest *_struct.Struct) (*_struct.Struct, error) {
	if existing == nil {
		return latest, nil
	}
	if latest == nil {
		return existing, nil
	}

	// To merge latest into existing we first create a patch object that consists of applying changes from latest to
	// an empty struct. Then we apply this patch to existing so that the values changed in latest take precedence but
	// barring conflicts/overwrites the values in existing stay the same.
	jsonExisting, err := protojson.Marshal(existing)
	if err != nil {
		return nil, err
	}
	jsonLatest, err := protojson.Marshal(latest)
	if err != nil {
		return nil, err
	}
	patch, err := jsonpatch.CreateMergePatch(jsonEmpty, jsonLatest)
	if err != nil {
		return nil, err
	}
	custom, err := jsonpatch.MergePatch(jsonExisting, patch)
	if err != nil {
		return nil, err
	}
	var response _struct.Struct

	err = protojson.Unmarshal(custom, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// mergeExternalResource combines the lastest ExternalResourceInfo proto with an existing instance
// by updating fields and merging logs.
func mergeExternalResource(existing, latest *event.ExternalResourceInfo) *event.ExternalResourceInfo {
	if existing == nil {
		return latest
	}

	if latest == nil {
		return existing
	}

	if latest.ExternalId != "" && existing.ExternalId != latest.ExternalId {
		existing.ExternalId = latest.ExternalId
	}
	// we only update if the index is equal so updating existing.Index is not necessary
	if latest.RetryAttempt != 0 && existing.RetryAttempt != latest.RetryAttempt {
		existing.RetryAttempt = latest.RetryAttempt
	}
	existing.Phase = latest.Phase
	if latest.CacheStatus != core.CatalogCacheStatus_CACHE_DISABLED && existing.CacheStatus != latest.CacheStatus {
		existing.CacheStatus = latest.CacheStatus
	}
	existing.Logs = mergeLogs(existing.Logs, latest.Logs)

	return existing
}

// mergeExternalResources combines lists of external resources. This involves appending new
// resources and updating in-place resources attributes.
func mergeExternalResources(existing, latest []*event.ExternalResourceInfo) []*event.ExternalResourceInfo {
	if len(latest) == 0 {
		return existing
	}

	for i, externalResource := range latest {
		// to ensure we update the correct ExternalResource we identify the resource index based on
		// one of two categories:
		// (1) a simple ID of an ExternalResource will only have the ExternalID populated. therefore
		// we use the index into the event array (ie. i).
		// (2) a subtask which contains an index, log links, phase, etc. if the index is set
		// (ie. != 0) or if any of the other fields are set we use the ExternalResource index.
		//
		// therefore, if we want to track any fields (other than ExternalID) we need to provide an
		// index to ensure correctness. additionally, if we only track ExternalID, any additions
		// need to include all previous ExternalResources.
		var index int
		if externalResource.GetIndex() == 0 && externalResource.GetCacheStatus() == 0 && len(externalResource.GetLogs()) != 0 &&
			externalResource.GetPhase() == 0 && externalResource.GetRetryAttempt() == 0 {
			index = i
		} else {
			index = int(externalResource.GetIndex())
		}

		if index >= len(existing) {
			// if the latest external resources contains an out of order update
			// (ie. index > len(existing)) then we should append placeholder external resources
			// that will be updated later
			for index-1 >= len(existing) {
				existing = append(existing, &event.ExternalResourceInfo{Index: uint32(len(existing))})
			}

			existing = append(existing, externalResource)
		} else {
			existing[index] = mergeExternalResource(existing[index], externalResource)
		}
	}

	return existing
}

// mergeMetadata merges an existing TaskExecutionMetadata instance with the provided instance. This
// includes updating non-defaulted fields and merging ExternalResources.
func mergeMetadata(existing, latest *event.TaskExecutionMetadata) *event.TaskExecutionMetadata {
	if existing == nil {
		return latest
	}

	if latest == nil {
		return existing
	}

	if latest.GeneratedName != "" && existing.GeneratedName != latest.GeneratedName {
		existing.GeneratedName = latest.GeneratedName
	}
	existing.ExternalResources = mergeExternalResources(existing.ExternalResources, latest.ExternalResources)
	existing.ResourcePoolInfo = latest.ResourcePoolInfo
	if latest.PluginIdentifier != "" && existing.PluginIdentifier != latest.PluginIdentifier {
		existing.PluginIdentifier = latest.PluginIdentifier
	}
	if latest.InstanceClass != event.TaskExecutionMetadata_DEFAULT && existing.InstanceClass != latest.InstanceClass {
		existing.InstanceClass = latest.InstanceClass
	}

	return existing
}

func UpdateTaskExecutionModel(ctx context.Context, request *admin.TaskExecutionEventRequest, taskExecutionModel *models.TaskExecution,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	var taskExecutionClosure admin.TaskExecutionClosure
	err := proto.Unmarshal(taskExecutionModel.Closure, &taskExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to unmarshal task execution closure with error: %+v", err)
	}
	existingTaskPhase := taskExecutionModel.Phase
	taskExecutionModel.Phase = request.Event.Phase.String()
	taskExecutionModel.PhaseVersion = request.Event.PhaseVersion
	taskExecutionClosure.Phase = request.Event.Phase
	taskExecutionClosure.UpdatedAt = request.Event.OccurredAt
	taskExecutionClosure.Logs = mergeLogs(taskExecutionClosure.Logs, request.Event.Logs)
	if len(request.Event.Reason) > 0 {
		taskExecutionClosure.Reason = request.Event.Reason
	}
	if existingTaskPhase != core.TaskExecution_RUNNING.String() && taskExecutionModel.Phase == core.TaskExecution_RUNNING.String() {
		err = addTaskStartedState(request, taskExecutionModel, &taskExecutionClosure)
		if err != nil {
			return err
		}
	}

	if common.IsTaskExecutionTerminal(request.Event.Phase) {
		err := addTaskTerminalState(ctx, request, taskExecutionModel, &taskExecutionClosure, inlineEventDataPolicy, storageClient)
		if err != nil {
			return err
		}
	}
	taskExecutionClosure.CustomInfo, err = mergeCustom(taskExecutionClosure.CustomInfo, request.Event.CustomInfo)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to merge task event custom_info with error: %v", err)
	}
	taskExecutionClosure.Metadata = mergeMetadata(taskExecutionClosure.Metadata, request.Event.Metadata)
	marshaledClosure, err := proto.Marshal(&taskExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal task execution closure with error: %v", err)
	}
	taskExecutionModel.Closure = marshaledClosure
	updatedAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to parse updated at timestamp")
	}
	taskExecutionModel.TaskExecutionUpdatedAt = &updatedAt
	return nil
}

func FromTaskExecutionModel(taskExecutionModel models.TaskExecution) (*admin.TaskExecution, error) {
	var closure admin.TaskExecutionClosure
	err := proto.Unmarshal(taskExecutionModel.Closure, &closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}

	taskExecution := &admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      taskExecutionModel.TaskExecutionKey.TaskKey.Project,
				Domain:       taskExecutionModel.TaskExecutionKey.TaskKey.Domain,
				Name:         taskExecutionModel.TaskExecutionKey.TaskKey.Name,
				Version:      taskExecutionModel.TaskExecutionKey.TaskKey.Version,
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: taskExecutionModel.NodeExecutionKey.NodeID,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Project,
					Domain:  taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Domain,
					Name:    taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Name,
				},
			},
			RetryAttempt: *taskExecutionModel.TaskExecutionKey.RetryAttempt,
		},
		InputUri: taskExecutionModel.InputURI,
		Closure:  &closure,
	}
	if len(taskExecutionModel.ChildNodeExecution) > 0 {
		taskExecution.IsParent = true
	}

	return taskExecution, nil
}

func FromTaskExecutionModels(taskExecutionModels []models.TaskExecution) ([]*admin.TaskExecution, error) {
	taskExecutions := make([]*admin.TaskExecution, len(taskExecutionModels))
	for idx, taskExecutionModel := range taskExecutionModels {
		taskExecution, err := FromTaskExecutionModel(taskExecutionModel)
		if err != nil {
			return nil, err
		}
		taskExecutions[idx] = taskExecution
	}
	return taskExecutions, nil
}
