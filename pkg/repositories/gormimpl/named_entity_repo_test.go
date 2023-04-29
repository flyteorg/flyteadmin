package gormimpl

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/flyteorg/flytestdlib/promutils"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func getMockNamedEntityResponseFromDb(expected models.NamedEntity) map[string]interface{} {
	metadata := make(map[string]interface{})
	metadata["resource_type"] = expected.ResourceType
	metadata["project"] = expected.Project
	metadata["domain"] = expected.Domain
	metadata["name"] = expected.Name
	metadata["description"] = expected.Description
	metadata["state"] = expected.State
	return metadata
}

func TestGetNamedEntity(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	results := make([]map[string]interface{}, 0)
	metadata := getMockNamedEntityResponseFromDb(models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: description,
		},
	})
	results = append(results, metadata)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(
		`SELECT workflows.project,workflows.domain,workflows.name,'2' AS resource_type,named_entity_metadata.description,named_entity_metadata.state FROM "workflows" LEFT JOIN named_entity_metadata ON named_entity_metadata.resource_type = 2 AND named_entity_metadata.project = workflows.project AND named_entity_metadata.domain = workflows.domain AND named_entity_metadata.name = workflows.name WHERE workflows.project = $1 AND workflows.domain = $2 AND workflows.name = $3 LIMIT 1`).WithReply(results)
	output, err := metadataRepo.Get(context.Background(), interfaces.GetNamedEntityInput{
		ResourceType: resourceType,
		Project:      project,
		Domain:       domain,
		Name:         name,
	})
	assert.NoError(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, name, output.Name)
	assert.Equal(t, resourceType, output.ResourceType)
	assert.Equal(t, description, output.Description)
}

func TestUpdateNamedEntity_WithExisting(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	const updatedDescription = "updated description"

	results := make([]map[string]interface{}, 0)
	activeState := int32(admin.NamedEntityState_NAMED_ENTITY_ACTIVE)
	metadata := getMockNamedEntityResponseFromDb(models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: description,
			State:       &activeState,
		},
	})
	results = append(results, metadata)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(
		`SELECT "named_entity_metadata"."created_at","named_entity_metadata"."updated_at","named_entity_metadata"."deleted_at","named_entity_metadata"."resource_type","named_entity_metadata"."project","named_entity_metadata"."domain","named_entity_metadata"."name","named_entity_metadata"."description","named_entity_metadata"."state" FROM "named_entity_metadata" WHERE "named_entity_metadata"."resource_type" = $1 AND "named_entity_metadata"."project" = $2 AND "named_entity_metadata"."domain" = $3 AND "named_entity_metadata"."name" = $4 ORDER BY "named_entity_metadata"."id" LIMIT 1`).WithReply(results)

	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(
		`UPDATE "named_entity_metadata" SET "description"=$1,"state"=$2,"updated_at"=$3 WHERE "named_entity_metadata"."resource_type" = $4 AND "named_entity_metadata"."project" = $5 AND "named_entity_metadata"."domain" = $6 AND "named_entity_metadata"."name" = $7 AND "resource_type" = $8 AND "project" = $9 AND "domain" = $10 AND "name" = $11`)

	err := metadataRepo.Update(context.Background(), models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: updatedDescription,
			State:       &activeState,
		},
	})
	assert.NoError(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestUpdateNamedEntity_CreateNew(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	const updatedDescription = "updated description"

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(
		`INSERT INTO "named_entity_metadata" ("created_at","updated_at","deleted_at","resource_type","project","domain","name","description","state") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`)

	err := metadataRepo.Update(context.Background(), models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: updatedDescription,
		},
	})
	assert.NoError(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestListNamedEntity(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	results := make([]map[string]interface{}, 0)
	metadata := getMockNamedEntityResponseFromDb(models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: resourceType,
			Project:      project,
			Domain:       domain,
			Name:         name,
		},
		NamedEntityMetadataFields: models.NamedEntityMetadataFields{
			Description: description,
		},
	})
	results = append(results, metadata)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockQuery := GlobalMock.NewMock()

	mockQuery.WithQuery(
		`SELECT entities.project,entities.domain,entities.name,'2' AS resource_type,named_entity_metadata.description,named_entity_metadata.state FROM "named_entity_metadata" RIGHT JOIN (SELECT project,domain,name FROM "workflows" WHERE "domain" = $1 AND "project" = $2 GROUP BY project, domain, name, created_at ORDER BY name desc LIMIT 20) AS entities ON named_entity_metadata.resource_type = 2 AND named_entity_metadata.project = entities.project AND named_entity_metadata.domain = entities.domain AND named_entity_metadata.name = entities.name GROUP BY entities.project, entities.domain, entities.name, named_entity_metadata.description, named_entity_metadata.state, named_entity_metadata.created_at ORDER BY name desc`).WithReply(results)

	sortParameter, _ := common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "name",
	})
	output, err := metadataRepo.List(context.Background(), interfaces.ListNamedEntityInput{
		ResourceType: resourceType,
		Project:      "admintests",
		Domain:       "development",
		ListResourceInput: interfaces.ListResourceInput{
			Limit:         20,
			SortParameter: sortParameter,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, output.Entities, 1)
}

func TestListNamedEntityTxErrorCases(t *testing.T) {
	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	mockQuery := GlobalMock.NewMock()

	mockQuery.WithQuery(
		`SELECT entities.project,entities.domain,entities.name,'2' AS resource_type,named_entity_metadata.description,named_entity_metadata.state FROM "named_entity_metadata" RIGHT JOIN (SELECT project,domain,name FROM "workflows" WHERE "domain" = $1 AND "project" = $2 GROUP BY project, domain, name, created_at ORDER BY name desc LIMIT 20) AS entities ON named_entity_metadata.resource_type = 2 AND named_entity_metadata.project = entities.project AND named_entity_metadata.domain = entities.domain AND named_entity_metadata.name = entities.name GROUP BY entities.project, entities.domain, entities.name, named_entity_metadata.description, named_entity_metadata.state, named_entity_metadata.created_at ORDER BY name desc%`).WithError(fmt.Errorf("failed"))

	sortParameter, _ := common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "name",
	})
	output, err := metadataRepo.List(context.Background(), interfaces.ListNamedEntityInput{
		ResourceType: resourceType,
		Project:      "admintests",
		Domain:       "development",
		ListResourceInput: interfaces.ListResourceInput{
			Limit:         20,
			SortParameter: sortParameter,
		},
	})
	assert.Equal(t, "Test transformer failed to find transformation to apply", err.Error())
	assert.Len(t, output.Entities, 0)
}

func TestListNamedEntityInputErrorCases(t *testing.T) {
	type test struct {
		input        interfaces.ListNamedEntityInput
		wantedError  error
		wantedLength int
	}

	sortParameter, _ := common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "name",
	})

	tests := []test{
		{
			input: interfaces.ListNamedEntityInput{
				ResourceType: resourceType,
				Project:      "",
				Domain:       "development",
				ListResourceInput: interfaces.ListResourceInput{
					Limit:         20,
					SortParameter: sortParameter,
				},
			},
			wantedError:  errors.GetInvalidInputError(Project),
			wantedLength: 0,
		},
		{
			input: interfaces.ListNamedEntityInput{
				ResourceType: resourceType,
				Project:      "project",
				Domain:       "",
				ListResourceInput: interfaces.ListResourceInput{
					Limit:         20,
					SortParameter: sortParameter,
				},
			},
			wantedError:  errors.GetInvalidInputError(Domain),
			wantedLength: 0,
		},
		{
			input: interfaces.ListNamedEntityInput{
				ResourceType: resourceType,
				Project:      "project",
				Domain:       "development",
				ListResourceInput: interfaces.ListResourceInput{
					Limit:         0,
					SortParameter: sortParameter,
				},
			},
			wantedError:  errors.GetInvalidInputError(limit),
			wantedLength: 0,
		},
		{
			input: interfaces.ListNamedEntityInput{
				ResourceType: -1,
				Project:      "project",
				Domain:       "development",
				ListResourceInput: interfaces.ListResourceInput{
					Limit:         20,
					SortParameter: sortParameter,
				},
			},
			wantedError: adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"Cannot list entity names for resource type: %v", -1),
			wantedLength: 0,
		},
	}

	metadataRepo := NewNamedEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	for _, tc := range tests {
		output, err := metadataRepo.List(context.Background(), tc.input)
		assert.Len(t, output.Entities, tc.wantedLength)
		assert.Equal(t, tc.wantedError, err)
	}
}
