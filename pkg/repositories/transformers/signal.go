package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
)

func CreateSignalModel(signal *admin.Signal) models.Signal {
	valueBytes, err := proto.Marshal(signal.Value)
	if err != nil {
		return models.Signal{}
	}

	return models.Signal{
		SignalKey: models.SignalKey{
			ExecutionKey: models.ExecutionKey{
				Project: signal.Id.ExecutionId.Project,
				Domain:  signal.Id.ExecutionId.Domain,
				Name:    signal.Id.ExecutionId.Name,
			},
			SignalID: signal.Id.SignalId,
		},
		Value: valueBytes,
	}
}

func FromSignalModel(signalModel models.Signal) admin.Signal {
	valueDeserialized := &core.Literal{}
	err := proto.Unmarshal(signalModel.Value, valueDeserialized)
	if err != nil {
		return admin.Signal{}
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
	}
}
