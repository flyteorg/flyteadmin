package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/golang/protobuf/proto"

	"google.golang.org/grpc/codes"
)

func CreateSignalModel(id core.SignalIdentifier, signalType *core.LiteralType, value *core.Literal) (models.Signal, error) {
	var typeBytes []byte
	var err error
	if signalType != nil {
		typeBytes, err = proto.Marshal(signalType)
		if err != nil {
			return models.Signal{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize signal type")
		}
	}

	var valueBytes []byte
	if value != nil {
		valueBytes, err = proto.Marshal(value)
		if err != nil {
			return models.Signal{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize signal value")
		}
	}

	return models.Signal{
		SignalKey: models.SignalKey{
			ExecutionKey: models.ExecutionKey{
				Project: id.ExecutionId.Project,
				Domain:  id.ExecutionId.Domain,
				Name:    id.ExecutionId.Name,
			},
			SignalID: id.SignalId,
		},
		Type:  typeBytes,
		Value: valueBytes,
	}, nil
}

func FromSignalModel(signalModel models.Signal) (admin.Signal, error) {
	var typeDeserialized core.LiteralType
	if len(signalModel.Type) > 0 {
		err := proto.Unmarshal(signalModel.Type, &typeDeserialized)
		if err != nil {
			return admin.Signal{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal signal type")
		}
	}

	var valueDeserialized core.Literal
	if len(signalModel.Value) > 0 {
		err := proto.Unmarshal(signalModel.Value, &valueDeserialized)
		if err != nil {
			return admin.Signal{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal signal value")
		}
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
		Type:  &typeDeserialized,
		Value: &valueDeserialized,
	}, nil
}
