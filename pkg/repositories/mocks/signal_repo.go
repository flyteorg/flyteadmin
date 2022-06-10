// Mock implementation of a task repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type GetSignalFunc func(input models.SignalKey) (models.Signal, error)
type GetOrCreateSignalFunc func(input *models.Signal) error
type ListSignalsFunc func(input interfaces.ListResourceInput) ([]models.Signal, error)
type UpdateSignalFunc func(input models.SignalKey, value []byte) error

type MockSignalRepo struct {
	getFunction         GetSignalFunc
	getOrCreateFunction GetOrCreateSignalFunc
	listFunction        ListSignalsFunc
	updateFunction      UpdateSignalFunc
}

func (r *MockSignalRepo) Get(ctx context.Context, input models.SignalKey) (models.Signal, error) {
	if r.getFunction != nil {
		return r.getFunction(input)
	}
	return models.Signal{}, nil
}

func (r *MockSignalRepo) SetGetCallback(getFunction GetSignalFunc) {
	r.getFunction = getFunction
}

func (r *MockSignalRepo) GetOrCreate(ctx context.Context, input *models.Signal) error {
	if r.getOrCreateFunction != nil {
		return r.getOrCreateFunction(input)
	}
	return nil
}

func (r *MockSignalRepo) SetGetOrCreateCallback(getOrCreateFunction GetOrCreateSignalFunc) {
	r.getOrCreateFunction = getOrCreateFunction
}

func (r *MockSignalRepo) List(ctx context.Context, input interfaces.ListResourceInput) ([]models.Signal, error) {
	if r.listFunction != nil {
		return r.listFunction(input)
	}
	return nil, nil
}

func (r *MockSignalRepo) SetListCallback(listFunction ListSignalsFunc) {
	r.listFunction = listFunction
}

func (r *MockSignalRepo) Update(ctx context.Context, input models.SignalKey, value []byte) error {
	if r.updateFunction != nil {
		return r.updateFunction(input, value)
	}
	return nil
}

func (r *MockSignalRepo) SetUpdateCallback(updateFunction UpdateSignalFunc) {
	r.updateFunction = updateFunction
}

func NewMockSignalRepo() interfaces.SignalRepoInterface {
	return &MockSignalRepo{}
}
