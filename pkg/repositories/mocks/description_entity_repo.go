// Mock implementation of a workflow repo to be used for tests.
package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

type CreateDescriptionEntityFunc func(input models.DescriptionEntity) error
type GetDescriptionEntityFunc func(input models.DescriptionEntityKey) (models.DescriptionEntity, error)

type MockDescriptionEntityRepo struct {
	createFunction CreateDescriptionEntityFunc
	getFunction    GetDescriptionEntityFunc
}

func (r *MockDescriptionEntityRepo) Create(ctx context.Context, DescriptionEntity models.DescriptionEntity) error {
	if r.createFunction != nil {
		return r.createFunction(DescriptionEntity)
	}
	return nil
}

func (r *MockDescriptionEntityRepo) Get(
	ctx context.Context, input models.DescriptionEntityKey) (models.DescriptionEntity, error) {
	if r.getFunction != nil {
		return r.getFunction(input)
	}
	return models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			ResourceType: input.ResourceType,
			Project:      input.Project,
			Domain:       input.Domain,
			Name:         input.Name,
		},
		ShortDescription: "hello world",
	}, nil
}

func (r *MockDescriptionEntityRepo) SetCreateCallback(createFunction CreateDescriptionEntityFunc) {
	r.createFunction = createFunction
}

func (r *MockDescriptionEntityRepo) SetGetCallback(getFunction GetDescriptionEntityFunc) {
	r.getFunction = getFunction
}

func NewMockDescriptionEntityRepo() interfaces.DescriptionEntityRepoInterface {
	return &MockDescriptionEntityRepo{}
}
