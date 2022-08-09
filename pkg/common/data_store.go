package common

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	//"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	errrs "github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
)

func OffloadLiteralMap(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, nestedKeys ...string) (storage.DataReference, error) {
	if literalMap == nil {
		literalMap = &core.LiteralMap{}
	}
	return OffloadProto(ctx, storageClient, literalMap, nestedKeys...)
}

func OffloadProto(ctx context.Context, storageClient *storage.DataStore, msg proto.Message, nestedKeys ...string) (storage.DataReference, error) {
	return OffloadProtoWithRetryDelayAndAttempts(ctx, storageClient, msg, async.RetryDelay, 5, nestedKeys...)
}

func OffloadProtoWithRetryDelayAndAttempts(ctx context.Context, storageClient *storage.DataStore, msg proto.Message, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	nestedKeyReference := []string{
		shared.Metadata,
	}
	nestedKeyReference = append(nestedKeyReference, nestedKeys...)
	uri, err := storageClient.ConstructReference(ctx, storageClient.GetBaseContainerFQN(ctx), nestedKeyReference...)
	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to construct data reference for [%+v] with err: %v", nestedKeys, err)
	}

	err = async.RetryOnSpecificErrors(attempts, retryDelay, func() error {
		err = storageClient.WriteProtobuf(ctx, uri, storage.Options{}, msg)
		return err
	}, isRetryableError)

	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to write protobuf for [%+v] with err: %v", nestedKeys, err)
	}

	return uri, nil
}

func isRetryableError(err error) bool {
	if e, ok := errrs.Cause(err).(*googleapi.Error); ok && e.Code == 409 {
		return true
	}
	return false
}

/*func OffloadWorkflowClosure(ctx context.Context, storageClient *storage.DataStore, flyteWf *v1alpha1.FlyteWorkflow, workflowClosure *core.CompiledWorkflowClosure, executionID v1alpha1.ExecutionID) error {
	reference, err := store(ctx, storageClient, workflowClosure, nestedKeys(executionID, shared.WorkflowClosure)...)
	if err != nil {
		return err
	}

	flyteWf.WorkflowClosureDataReference = reference
	flyteWf.WorkflowSpec = nil
	flyteWf.SubWorkflows = nil
	flyteWf.Tasks = nil
	return nil
}

func nestedKeys(execID v1alpha1.ExecutionID, filename string) []string {
	return []string{shared.Metadata, execID.GetProject(), execID.Domain, execID.Name, filename}
}

func store(ctx context.Context, storageClient *storage.DataStore, workflowClosure *core.CompiledWorkflowClosure, nestedKeys ...string) (storage.DataReference, error) {

	base := storageClient.GetBaseContainerFQN(ctx)
	remoteClosureDataRef, err := storageClient.ConstructReference(ctx, base, nestedKeys...)
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}
	err = storageClient.WriteProtobuf(ctx, remoteClosureDataRef, storage.Options{}, workflowClosure)
	if err != nil {
		return "", err
	}

	return remoteClosureDataRef, nil
}*/
