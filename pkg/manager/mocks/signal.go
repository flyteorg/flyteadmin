package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type GetOrCreateSignalFunc func(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error)
type ListSignalsFunc func(ctx context.Context, request admin.SignalListRequest) ([]*admin.Signal, error)
type SetSignalFunc func(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error)

type MockSignalManager struct {
	getOrCreateSignalFunc GetOrCreateSignalFunc
	listSignalsFunc       ListSignalsFunc
	setSignalFunc         SetSignalFunc
}

func (r *MockSignalManager) SetGetOrCreateCallback(getOrCreateFunction GetOrCreateSignalFunc) {
	r.getOrCreateSignalFunc = getOrCreateFunction
}

func (r *MockSignalManager) GetOrCreateSignal(
	ctx context.Context,
	request admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
	if r.getOrCreateSignalFunc != nil {
		return r.getOrCreateSignalFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockSignalManager) SetListCallback(listFunction ListSignalsFunc) {
	r.listSignalsFunc = listFunction
}

func (r *MockSignalManager) ListSignals(
	ctx context.Context,
	request admin.SignalListRequest) ([]*admin.Signal, error) {
	if r.listSignalsFunc != nil {
		return r.listSignalsFunc(ctx, request)
	}
	return nil, nil
}

func (r *MockSignalManager) SetSetCallback(setFunction SetSignalFunc) {
	r.setSignalFunc = setFunction
}

func (r *MockSignalManager) SetSignal(
	ctx context.Context,
	request admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
	if r.setSignalFunc != nil {
		return r.setSignalFunc(ctx, request)
	}
	return nil, nil
}
