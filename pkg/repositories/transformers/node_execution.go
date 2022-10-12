package transformers

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	genModel "github.com/flyteorg/flyteadmin/pkg/repositories/gen/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
)

type ToNodeExecutionModelInput struct {
	Request                      *admin.NodeExecutionEventRequest
	ParentTaskExecutionID        *uint
	ParentID                     *uint
	DynamicWorkflowRemoteClosure string
	InlineEventDataPolicy        interfaces.InlineEventDataPolicy
	StorageClient                *storage.DataStore
}

func addNodeRunningState(request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	closure *admin.NodeExecutionClosure) error {
	occurredAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal occurredAt with error: %v", err)
	}

	nodeExecutionModel.StartedAt = &occurredAt
	startedAtProto, err := ptypes.TimestampProto(occurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to marshal occurredAt into a timestamp proto with error: %v", err)
	}
	closure.StartedAt = startedAtProto
	return nil
}

func addTerminalState(
	ctx context.Context,
	request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	closure *admin.NodeExecutionClosure, inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	if closure.StartedAt == nil {
		logger.Warning(context.Background(), "node execution is missing StartedAt")
	} else {
		endTime, err := ptypes.Timestamp(request.Event.OccurredAt)
		if err != nil {
			return errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to parse node execution occurred at timestamp: %v", err)
		}
		nodeExecutionModel.Duration = endTime.Sub(*nodeExecutionModel.StartedAt)
		closure.Duration = ptypes.DurationProto(nodeExecutionModel.Duration)
	}

	// Serialize output results (if they exist)
	if request.Event.GetOutputUri() != "" {
		closure.OutputResult = &admin.NodeExecutionClosure_OutputUri{
			OutputUri: request.Event.GetOutputUri(),
		}
	} else if request.Event.GetOutputData() != nil {
		switch inlineEventDataPolicy {
		case interfaces.InlineEventDataPolicyStoreInline:
			closure.OutputResult = &admin.NodeExecutionClosure_OutputData{
				OutputData: request.Event.GetOutputData(),
			}
		default:
			logger.Debugf(ctx, "Offloading outputs per InlineEventDataPolicy")
			uri, err := common.OffloadLiteralMap(ctx, storageClient, request.Event.GetOutputData(),
				request.Event.Id.ExecutionId.Project, request.Event.Id.ExecutionId.Domain, request.Event.Id.ExecutionId.Name,
				request.Event.Id.NodeId, OutputsObjectSuffix)
			if err != nil {
				return err
			}
			closure.OutputResult = &admin.NodeExecutionClosure_OutputUri{
				OutputUri: uri.String(),
			}
		}
	} else if request.Event.GetError() != nil {
		closure.OutputResult = &admin.NodeExecutionClosure_Error{
			Error: request.Event.GetError(),
		}
		k := request.Event.GetError().Kind.String()
		nodeExecutionModel.ErrorKind = &k
		nodeExecutionModel.ErrorCode = &request.Event.GetError().Code
	}
	closure.DeckUri = request.Event.DeckUri

	return nil
}

func CreateNodeExecutionModel(ctx context.Context, input ToNodeExecutionModelInput) (*models.NodeExecution, error) {
	nodeExecution := &models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: input.Request.Event.Id.NodeId,
			ExecutionKey: models.ExecutionKey{
				Project: input.Request.Event.Id.ExecutionId.Project,
				Domain:  input.Request.Event.Id.ExecutionId.Domain,
				Name:    input.Request.Event.Id.ExecutionId.Name,
			},
		},
		Phase:    input.Request.Event.Phase.String(),
		InputURI: input.Request.Event.InputUri,
	}

	closure := admin.NodeExecutionClosure{
		Phase:     input.Request.Event.Phase,
		CreatedAt: input.Request.Event.OccurredAt,
		UpdatedAt: input.Request.Event.OccurredAt,
		DeckUri:   input.Request.Event.DeckUri,
	}

	nodeExecutionMetadata := admin.NodeExecutionMetaData{
		RetryGroup:   input.Request.Event.RetryGroup,
		SpecNodeId:   input.Request.Event.SpecNodeId,
		IsParentNode: input.Request.Event.IsParent,
		IsDynamic:    input.Request.Event.IsDynamic,
	}

	if input.Request.Event.Phase == core.NodeExecution_RUNNING {
		err := addNodeRunningState(input.Request, nodeExecution, &closure)
		if err != nil {
			return nil, err
		}
	}
	if common.IsNodeExecutionTerminal(input.Request.Event.Phase) {
		err := addTerminalState(ctx, input.Request, nodeExecution, &closure, input.InlineEventDataPolicy, input.StorageClient)
		if err != nil {
			return nil, err
		}
	}
	marshaledClosure, err := proto.Marshal(&closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal node execution closure with error: %v", err)
	}
	marshaledNodeExecutionMetadata, err := proto.Marshal(&nodeExecutionMetadata)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal node execution metadata with error: %v", err)
	}
	nodeExecution.Closure = marshaledClosure
	nodeExecution.NodeExecutionMetadata = marshaledNodeExecutionMetadata
	nodeExecutionCreatedAt, err := ptypes.Timestamp(input.Request.Event.OccurredAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read event timestamp")
	}
	nodeExecution.NodeExecutionCreatedAt = &nodeExecutionCreatedAt
	nodeExecution.NodeExecutionUpdatedAt = &nodeExecutionCreatedAt
	if input.Request.Event.ParentTaskMetadata != nil {
		nodeExecution.ParentTaskExecutionID = input.ParentTaskExecutionID
	}
	nodeExecution.ParentID = input.ParentID
	nodeExecution.DynamicWorkflowRemoteClosureReference = input.DynamicWorkflowRemoteClosure

	internalData := &genModel.NodeExecutionInternalData{
		EventVersion: input.Request.Event.EventVersion,
	}
	internalDataBytes, err := proto.Marshal(internalData)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to marshal node execution data with err: %v", err)
	}
	nodeExecution.InternalData = internalDataBytes
	return nodeExecution, nil
}

func UpdateNodeExecutionModel(
	ctx context.Context, request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	targetExecution *core.WorkflowExecutionIdentifier, dynamicWorkflowRemoteClosure string,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	var nodeExecutionClosure admin.NodeExecutionClosure
	err := proto.Unmarshal(nodeExecutionModel.Closure, &nodeExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to unmarshal node execution closure with error: %+v", err)
	}
	nodeExecutionModel.Phase = request.Event.Phase.String()
	nodeExecutionClosure.Phase = request.Event.Phase
	nodeExecutionClosure.UpdatedAt = request.Event.OccurredAt

	if request.Event.Phase == core.NodeExecution_RUNNING {
		err := addNodeRunningState(request, nodeExecutionModel, &nodeExecutionClosure)
		if err != nil {
			return err
		}
	}
	if common.IsNodeExecutionTerminal(request.Event.Phase) {
		err := addTerminalState(ctx, request, nodeExecutionModel, &nodeExecutionClosure, inlineEventDataPolicy, storageClient)
		if err != nil {
			return err
		}
	}

	// If the node execution kicked off a workflow execution update the closure if it wasn't set
	if targetExecution != nil && nodeExecutionClosure.GetWorkflowNodeMetadata() == nil {
		nodeExecutionClosure.TargetMetadata = &admin.NodeExecutionClosure_WorkflowNodeMetadata{
			WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
				ExecutionId: targetExecution,
			},
		}
	}

	// Update TaskNodeMetadata, which includes caching information today.
	if request.Event.GetTaskNodeMetadata() != nil && request.Event.GetTaskNodeMetadata().CatalogKey != nil {
		st := request.Event.GetTaskNodeMetadata().GetCacheStatus().String()
		targetMetadata := &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{
				CacheStatus: request.Event.GetTaskNodeMetadata().GetCacheStatus(),
				CatalogKey:  request.Event.GetTaskNodeMetadata().GetCatalogKey(),
			},
		}
		nodeExecutionClosure.TargetMetadata = targetMetadata
		nodeExecutionModel.CacheStatus = &st
	}

	marshaledClosure, err := proto.Marshal(&nodeExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal node execution closure with error: %v", err)
	}

	nodeExecutionModel.Closure = marshaledClosure
	updatedAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to parse updated at timestamp")
	}
	nodeExecutionModel.NodeExecutionUpdatedAt = &updatedAt
	nodeExecutionModel.DynamicWorkflowRemoteClosureReference = dynamicWorkflowRemoteClosure

	// In the case of dynamic nodes reporting DYNAMIC_RUNNING, the IsParent and IsDynamic bits will be set for this event.
	// Update the node execution metadata accordingly.
	if request.Event.IsParent || request.Event.IsDynamic {
		var nodeExecutionMetadata admin.NodeExecutionMetaData
		if len(nodeExecutionModel.NodeExecutionMetadata) > 0 {
			if err := proto.Unmarshal(nodeExecutionModel.NodeExecutionMetadata, &nodeExecutionMetadata); err != nil {
				return errors.NewFlyteAdminErrorf(codes.Internal,
					"failed to unmarshal node execution metadata with error: %+v", err)
			}
		}
		// Not every event sends IsParent and IsDynamic as an artifact of how propeller handles dynamic nodes.
		// Only explicitly set the fields, when they're set in the event itself.
		if request.Event.IsParent {
			nodeExecutionMetadata.IsParentNode = true
		}
		if request.Event.IsDynamic {
			nodeExecutionMetadata.IsDynamic = true
		}
		nodeExecMetadataBytes, err := proto.Marshal(&nodeExecutionMetadata)
		if err != nil {
			return errors.NewFlyteAdminErrorf(codes.Internal,
				"failed to marshal node execution metadata with error: %+v", err)
		}
		nodeExecutionModel.NodeExecutionMetadata = nodeExecMetadataBytes
	}

	return nil
}

func FromNodeExecutionModel(nodeExecutionModel models.NodeExecution) (*admin.NodeExecution, error) {
	var closure admin.NodeExecutionClosure
	err := proto.Unmarshal(nodeExecutionModel.Closure, &closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}

	var nodeExecutionMetadata admin.NodeExecutionMetaData
	err = proto.Unmarshal(nodeExecutionModel.NodeExecutionMetadata, &nodeExecutionMetadata)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal nodeExecutionMetadata")
	}
	// TODO: delete this block and references to preloading child node executions no earlier than Q3 2022
	// This is required for historical reasons because propeller did not always send IsParent or IsDynamic in events.
	if !(nodeExecutionMetadata.IsParentNode || nodeExecutionMetadata.IsDynamic) {
		if len(nodeExecutionModel.ChildNodeExecutions) > 0 {
			nodeExecutionMetadata.IsParentNode = true
			if len(nodeExecutionModel.DynamicWorkflowRemoteClosureReference) > 0 {
				nodeExecutionMetadata.IsDynamic = true
			}
		}
	}

	return &admin.NodeExecution{
		Id: &core.NodeExecutionIdentifier{
			NodeId: nodeExecutionModel.NodeID,
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: nodeExecutionModel.NodeExecutionKey.ExecutionKey.Project,
				Domain:  nodeExecutionModel.NodeExecutionKey.ExecutionKey.Domain,
				Name:    nodeExecutionModel.NodeExecutionKey.ExecutionKey.Name,
			},
		},
		InputUri: nodeExecutionModel.InputURI,
		Closure:  &closure,
		Metadata: &nodeExecutionMetadata,
	}, nil
}

func GetNodeExecutionInternalData(internalData []byte) (*genModel.NodeExecutionInternalData, error) {
	var nodeExecutionInternalData genModel.NodeExecutionInternalData
	if len(internalData) > 0 {
		err := proto.Unmarshal(internalData, &nodeExecutionInternalData)
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal node execution data: %v", err)
		}
	}
	return &nodeExecutionInternalData, nil
}
