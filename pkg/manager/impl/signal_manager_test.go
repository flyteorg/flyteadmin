package impl

import (
	"context"
	"errors"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/assert"
)

var (
	signalId = &core.SignalIdentifier{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		SignalId: "signal",
	}

	signalType = &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_BOOLEAN,
		},
	}

	signalValue = &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Boolean{
							Boolean: false,
						},
					},
				},
			},
		},
	}
)

func TestGetOrCreateSignal(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()

	t.Run("Happy", func(t *testing.T) {
		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalGetOrCreateRequest{
			Id: signalId,
			Type: signalType,
		}

		response, err := signalManager.GetOrCreateSignal(context.Background(), request)
		assert.NoError(t, err)

		assert.True(t, proto.Equal(&admin.Signal{
			Id: signalId,
			Type: signalType,
		}, response))
	})

	t.Run("ValidationError", func(t *testing.T) {
		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalGetOrCreateRequest{
			Type: signalType,
		}

		_, err := signalManager.GetOrCreateSignal(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mockRepository.SignalRepo().(*repositoryMocks.MockSignalRepo).SetGetOrCreateCallback(func(input *models.Signal) error {
			return errors.New("foo")
		})

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalGetOrCreateRequest{
			Id: signalId,
			Type: signalType,
		}

		_, err := signalManager.GetOrCreateSignal(context.Background(), request)
		assert.Error(t, err)
	})
}

func TestSetSignal(t *testing.T) {
	mockRepository := repositoryMocks.NewMockRepository()
	
	signalModel, err := transformers.CreateSignalModel(signalId, signalType, nil)
	assert.NoError(t, err)

	t.Run("Happy", func(t *testing.T) {
		mockRepository.SignalRepo().(*repositoryMocks.MockSignalRepo).SetGetCallback(func(input models.SignalKey) (models.Signal, error) {
			return signalModel, nil
		})

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Id:    signalId,
			Value: signalValue,
		}

		response, err := signalManager.SetSignal(context.Background(), request)
		assert.NoError(t, err)

		assert.True(t, proto.Equal(&admin.SignalSetResponse{}, response))
	})

	t.Run("ValidationError", func(t *testing.T) {
		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Value: signalValue,
		}

		_, err := signalManager.SetSignal(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("DBGetError", func(t *testing.T) {
		mockRepository.SignalRepo().(*repositoryMocks.MockSignalRepo).SetGetCallback(func(input models.SignalKey) (models.Signal, error) {
			return models.Signal{}, errors.New("foo")
		})

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Id:    signalId,
			Value: signalValue,
		}

		_, err := signalManager.SetSignal(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("DBUpdateError", func(t *testing.T) {
		mockRepository.SignalRepo().(*repositoryMocks.MockSignalRepo).SetGetCallback(func(input models.SignalKey) (models.Signal, error) {
			return signalModel, nil
		})
		mockRepository.SignalRepo().(*repositoryMocks.MockSignalRepo).SetUpdateCallback(func(input models.SignalKey, value []byte) error {
			return errors.New("foo")
		})

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Id:    signalId,
			Value: signalValue,
		}

		_, err := signalManager.SetSignal(context.Background(), request)
		assert.Error(t, err)
	})
}
