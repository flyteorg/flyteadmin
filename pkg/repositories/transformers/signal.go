package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/golang/protobuf/proto"

	"google.golang.org/grpc/codes"
)

func CreateSignalModel(signalId *core.SignalIdentifier, signalType *core.LiteralType, signalValue *core.Literal) (models.Signal, error) {
	signalModel := models.Signal{}
	if signalId != nil {
		signalKey := &signalModel.SignalKey
		if signalId.ExecutionId != nil {
			executionKey := &signalKey.ExecutionKey
			if signalId.ExecutionId.Project != "" {
				executionKey.Project = signalId.ExecutionId.Project
			}
			if signalId.ExecutionId.Domain != "" {
				executionKey.Domain = signalId.ExecutionId.Domain
			}
			if signalId.ExecutionId.Name != "" {
				executionKey.Name = signalId.ExecutionId.Name
			}
		}

		if signalId.SignalId != "" {
			signalKey.SignalID = signalId.SignalId
		}
	}

	if signalType != nil {
		typeBytes, err := proto.Marshal(signalType)
		if err != nil {
			return models.Signal{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize signal type")
		}

		signalModel.Type = typeBytes
	}

	if signalValue != nil {
		valueBytes, err := proto.Marshal(signalValue)
		if err != nil {
			return models.Signal{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize signal value")
		}

		signalModel.Value = valueBytes
	}

	return signalModel, nil
}

func initSignalIdentifier(id *core.SignalIdentifier) *core.SignalIdentifier {
	if id == nil {
		id = &core.SignalIdentifier{}
	}
	return id
}

func initWorkflowExecutionIdentifier(id *core.WorkflowExecutionIdentifier) *core.WorkflowExecutionIdentifier {
	if id == nil {
		return &core.WorkflowExecutionIdentifier{}
	}
	return id
}

func FromSignalModel(signalModel models.Signal) (admin.Signal, error) {
	signal := admin.Signal{}

	var executionId *core.WorkflowExecutionIdentifier
	if signalModel.SignalKey.ExecutionKey.Project != "" {
		executionId = initWorkflowExecutionIdentifier(executionId)
		executionId.Project = signalModel.SignalKey.ExecutionKey.Project
	}
	if signalModel.SignalKey.ExecutionKey.Domain != "" {
		executionId = initWorkflowExecutionIdentifier(executionId)
		executionId.Domain = signalModel.SignalKey.ExecutionKey.Domain
	}
	if signalModel.SignalKey.ExecutionKey.Name != "" {
		executionId = initWorkflowExecutionIdentifier(executionId)
		executionId.Name = signalModel.SignalKey.ExecutionKey.Name
	}

	var signalId *core.SignalIdentifier
	if executionId != nil {
		signalId = initSignalIdentifier(signalId)
		signalId.ExecutionId = executionId
	}
	if signalModel.SignalKey.SignalID != "" {
		signalId = initSignalIdentifier(signalId)
		signalId.SignalId = signalModel.SignalKey.SignalID
	}

	if signalId != nil {
		signal.Id = signalId
	}

	if len(signalModel.Type) > 0 {
		var typeDeserialized core.LiteralType
		err := proto.Unmarshal(signalModel.Type, &typeDeserialized)
		if err != nil {
			return admin.Signal{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal signal type")
		}
		signal.Type = &typeDeserialized
	}

	if len(signalModel.Value) > 0 {
		var valueDeserialized core.Literal
		err := proto.Unmarshal(signalModel.Value, &valueDeserialized)
		if err != nil {
			return admin.Signal{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal signal value")
		}
		signal.Value = &valueDeserialized
	}

	return signal, nil
}

func FromSignalModels(signalModels []models.Signal) ([]*admin.Signal, error) {
	signals := make([]*admin.Signal, len(signalModels))
	for idx, signalModel := range signalModels {
		signal, err := FromSignalModel(signalModel)
		if err != nil {
			return nil, err
		}
		signals[idx] = &signal
	}
	return signals, nil
}
