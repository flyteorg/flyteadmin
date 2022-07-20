package common

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/storage"
	errrs "github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
)

func OffloadLiteralMap(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, nestedKeys ...string) (storage.DataReference, error) {
	return OffloadLiteralMapWithRetryDelayAndAttempts(ctx, storageClient, literalMap, async.RetryDelay, 5, nestedKeys...)
}

func OffloadLiteralMapWithRetryDelayAndAttempts(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, retryDelay time.Duration, attempts int, nestedKeys ...string) (storage.DataReference, error) {
	if literalMap == nil {
		literalMap = &core.LiteralMap{}
	}
	nestedKeyReference := []string{
		shared.Metadata,
	}
	nestedKeyReference = append(nestedKeyReference, nestedKeys...)
	uri, err := storageClient.ConstructReference(ctx, storageClient.GetBaseContainerFQN(ctx), nestedKeyReference...)
	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to construct data reference for [%+v] with err: %v", nestedKeys, err)
	}

	err = async.RetryOnSpecificErrors(attempts, retryDelay, func() error {
		err = storageClient.WriteProtobuf(ctx, uri, storage.Options{}, literalMap)
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

func OffloadCrd(ctx context.Context, storageClient *storage.DataStore, flyteWf *v1alpha1.FlyteWorkflow) error {
	reference, err := store(ctx, storageClient, flyteWf.WorkflowSpec, nestedKeys(flyteWf.GetExecutionID(), shared.CrdWorkflowSpec)...)
	if err != nil {
		return err
	}
	flyteWf.WorkflowSpecDataReference = reference

	if len(flyteWf.SubWorkflows) > 0 {
		reference, err := store(ctx, storageClient, flyteWf.SubWorkflows, nestedKeys(flyteWf.GetExecutionID(), shared.CrdSubWorkflows)...)
		if err != nil {
			return err
		}
		flyteWf.SubWorkflowsDataReference = reference
	}

	if len(flyteWf.Tasks) > 0 {
		reference, err = store(ctx, storageClient, flyteWf.Tasks, nestedKeys(flyteWf.GetExecutionID(), shared.CrdTasks)...)
		if err != nil {
			return err
		}
		flyteWf.TasksDataReference = reference
	}
	return nil
}

func nestedKeys(execID v1alpha1.ExecutionID, filename string) []string {
	return []string{shared.Metadata, execID.GetProject(), execID.Domain, execID.Name, shared.Crd, filename}
}

func store(ctx context.Context, storageClient *storage.DataStore, dataObj any, nestedKeys ...string) (storage.DataReference, error) {
	data, err := json.Marshal(dataObj)
	if err != nil {
		return "", err
	}
	base := storageClient.GetBaseContainerFQN(ctx)
	reference, err := storageClient.ConstructReference(ctx, base, nestedKeys...)
	if err != nil {
		return "", err
	}
	err = storageClient.WriteRaw(ctx, reference, int64(len(data)), storage.Options{}, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return reference, nil
}
