package impl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

var descriptionEntityIdentifier = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      project,
	Domain:       domain,
	Name:         name,
	Version:      version,
}

var badDescriptionEntityIdentifier = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      project,
	Domain:       domain,
	Name:         "",
	Version:      version,
}

func getMockRepositoryForDETest() interfaces.Repository {
	return repositoryMocks.NewMockRepository()
}

func getMockConfigForDETest() runtimeInterfaces.Configuration {
	mockConfig := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, nil, nil, nil)
	return mockConfig
}

func TestDescriptionEntityManager_Create(t *testing.T) {
	repository := getMockRepositoryForDETest()
	manager := NewDescriptionEntityManager(repository, getMockConfigForDETest(), mockScope.NewTestScope())

	shortDescription := "hello world"
	getFunction := func(input models.DescriptionEntityKey) (models.DescriptionEntity, error) {
		return models.DescriptionEntity{}, errors.NewFlyteAdminErrorf(codes.NotFound, "NotFound")
	}
	repository.DescriptionEntityRepo().(*repositoryMocks.MockDescriptionEntityRepo).SetGetCallback(getFunction)
	descriptionEntity := admin.DescriptionEntity{
		ShortDescription: shortDescription,
	}
	response, err := manager.CreateDescriptionEntity(context.Background(), admin.DescriptionEntityCreateRequest{
		DescriptionEntity: &descriptionEntity,
		Id:                &descriptionEntityIdentifier,
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	getFunction = func(input models.DescriptionEntityKey) (models.DescriptionEntity, error) {
		return models.DescriptionEntity{}, nil
	}
	repository.DescriptionEntityRepo().(*repositoryMocks.MockDescriptionEntityRepo).SetGetCallback(getFunction)
	response, err = manager.CreateDescriptionEntity(context.Background(), admin.DescriptionEntityCreateRequest{
		DescriptionEntity: &descriptionEntity,
		Id:                &descriptionEntityIdentifier,
	})
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestDescriptionEntityManager_Get(t *testing.T) {
	repository := getMockRepositoryForDETest()
	manager := NewDescriptionEntityManager(repository, getMockConfigForDETest(), mockScope.NewTestScope())

	response, err := manager.GetDescriptionEntity(context.Background(), admin.ObjectGetRequest{
		Id: &descriptionEntityIdentifier,
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = manager.GetDescriptionEntity(context.Background(), admin.ObjectGetRequest{
		Id: &badDescriptionEntityIdentifier,
	})
	assert.Error(t, err)
	assert.Nil(t, response)
}
