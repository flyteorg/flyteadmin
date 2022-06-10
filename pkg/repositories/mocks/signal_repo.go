// Mock implementation of a task repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type GetOrCreateSignalFunc func(input *models.Signal) error
type ListSignalsFunc func(input models.Signal) ([]*models.Signal, error)
type UpdateSignalFunc func(input models.Signal) error

type MockSignalRepo struct {
	getOrCreateFunction GetOrCreateSignalFunc
	listFunction        ListSignalsFunc
	updateFunction      UpdateSignalFunc
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

func (r *MockSignalRepo) List(ctx context.Context, input models.Signal) ([]*models.Signal, error) {
	if r.listFunction != nil {
		return r.listFunction(input)
	}
	return nil, nil
}

func (r *MockSignalRepo) SetListCallback(listFunction ListSignalsFunc) {
	r.listFunction = listFunction
}

func (r *MockSignalRepo) Update(ctx context.Context, input models.Signal) error {
	if r.updateFunction != nil {
		return r.updateFunction(input)
	}
	return nil
}

func (r *MockSignalRepo) SetUpdateCallback(updateFunction UpdateSignalFunc) {
	r.updateFunction = updateFunction
}

func NewMockSignalRepo() interfaces.SignalRepoInterface {
	return &MockSignalRepo{}
}
