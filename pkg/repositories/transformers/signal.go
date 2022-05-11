package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/golang/protobuf/proto"

	"google.golang.org/grpc/codes"
)

func CreateSignalModel(request admin.SignalCreateRequest) (models.Signal, error) {
	valueBytes, err := proto.Marshal(request.Value)
	if err != nil {
		return models.Signal{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize signal value")
	}

	return models.Signal{
		SignalKey: models.SignalKey{
			ExecutionKey: models.ExecutionKey{
				Project: request.Id.ExecutionId.Project,
				Domain:  request.Id.ExecutionId.Domain,
				Name:    request.Id.ExecutionId.Name,
			},
			SignalID: request.Id.SignalId,
		},
		Value: valueBytes,
	}, nil
}

func FromSignalModel(signalModel models.Signal) (admin.Signal, error) {
	valueDeserialized := &core.Literal{}
	err := proto.Unmarshal(signalModel.Value, valueDeserialized)
	if err != nil {
		return admin.Signal{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal signal value")
	}

	return admin.Signal{
		Id: &core.SignalIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: signalModel.SignalKey.ExecutionKey.Project,
				Domain:  signalModel.SignalKey.ExecutionKey.Domain,
				Name:    signalModel.SignalKey.ExecutionKey.Name,
			},
			SignalId: signalModel.SignalKey.SignalID,
		},
		Value: valueDeserialized,
	}, nil
}
