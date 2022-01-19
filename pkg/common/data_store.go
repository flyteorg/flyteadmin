package common

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
)

func OffloadLiteralMap(ctx context.Context, storageClient *storage.DataStore, literalMap *core.LiteralMap, nestedKeys ...string) (storage.DataReference, error) {
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

	/*
	TODO: retry only on
	{"json":{…}, "level":"error",
	"msg":"Failed to write to the raw store [gs://flyte-production-storage/metadata/payment-risk-datasets/production/cz7vntanluzg2fck4yim/inputs]
	Error: Failed to write data [172b] to path [metadata/payment-risk-datasets/production/cz7vntanluzg2fck4yim/inputs].: googleapi:
	Error 409: The metadata for object "metadata/payment-risk-datasets/production/cz7vntanluzg2fck4yim/inputs" was edited during the operation.
	Please try again., conflict", "ts":"2021-12-27T05:33:19Z"}
	*/
	err = async.Retry(5, async.RetryDelay, func() error {
		err = storageClient.WriteProtobuf(ctx, uri, storage.Options{}, literalMap)
		return err
	})

	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal, "Failed to write protobuf for [%+v] with err: %v", nestedKeys, err)
	}

	return uri, nil
}
