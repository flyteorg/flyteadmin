package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

func TestGetOrCreateSignal(t *testing.T) {
	ctx := context.Background()
	mockSignalManager := mocks.MockSignalManager{}

	t.Run("Happy", func(t *testing.T) {
		mockSignalManager.SetGetOrCreateCallback(
			func(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
				return &admin.Signal{}, nil
			},
		)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &mockSignalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.GetOrCreateSignal(ctx, &admin.SignalGetOrCreateRequest{})
		assert.NoError(t, err)
	})

	t.Run("NilRequestError", func(t *testing.T) {
		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &mockSignalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.GetOrCreateSignal(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		mockSignalManager.SetGetOrCreateCallback(
			func(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
				return nil, errors.New("foo")
			},
		)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &mockSignalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.GetOrCreateSignal(ctx, &admin.SignalGetOrCreateRequest{})
		assert.Error(t, err)
	})
}

func TestSetSignal(t *testing.T) {
	ctx := context.Background()
	mockSignalManager := mocks.MockSignalManager{}

	t.Run("Happy", func(t *testing.T) {
		mockSignalManager.SetSetCallback(
			func(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
				return &admin.SignalSetResponse{}, nil
			},
		)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &mockSignalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, &admin.SignalSetRequest{})
		assert.NoError(t, err)
	})


	t.Run("NilRequestError", func(t *testing.T) {
		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &mockSignalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		mockSignalManager.SetSetCallback(
			func(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
				return nil, errors.New("foo")
			},
		)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &mockSignalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, &admin.SignalSetRequest{})
		assert.Error(t, err)
	})
}
