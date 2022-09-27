package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

const shortDescription = "hello"

func TestGetDescriptionEntity(t *testing.T) {
	descriptionEntityRepo := NewDescriptionEntityRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	descriptionEntities := make([]map[string]interface{}, 0)
	descriptionEntity := getMockDescriptionEntityResponseFromDb(version, []byte{1, 2})
	descriptionEntities = append(descriptionEntities, descriptionEntity)

	output, err := descriptionEntityRepo.Get(context.Background(), models.DescriptionEntityKey{
		ResourceType: resourceType,
		Project:      project,
		Domain:       domain,
		Name:         name,
		Version:      version,
	})
	assert.Empty(t, output)
	assert.EqualError(t, err, "Test transformer failed to find transformation to apply")

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT "description_entities"."resource_type","description_entities"."project","description_entities"."domain","description_entities"."name","description_entities"."version","description_entities"."id","description_entities"."created_at","description_entities"."updated_at","description_entities"."deleted_at","description_entities"."digest","description_entities"."short_description","description_entities"."long_description","description_entities"."link" FROM "description_entities" INNER JOIN workflows ON description_entities.project = workflows.project AND description_entities.domain = workflows.domain AND description_entities.id = workflows.description_id WHERE (workflows.project = $1) AND (workflows.domain = $2) AND (workflows.name = $3) AND (workflows.version = $4) LIMIT 1`).
		WithReply(descriptionEntities)
	output, err = descriptionEntityRepo.Get(context.Background(), models.DescriptionEntityKey{
		ResourceType: resourceType,
		Project:      project,
		Domain:       domain,
		Name:         name,
		Version:      version,
	})
	assert.Empty(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, name, output.Name)
	assert.Equal(t, version, output.Version)
	assert.Equal(t, []byte{1, 2}, output.Digest)
	assert.Equal(t, shortDescription, output.ShortDescription)
}

func getMockDescriptionEntityResponseFromDb(version string, digest []byte) map[string]interface{} {
	descriptionEntity := make(map[string]interface{})
	descriptionEntity["resource_type"] = resourceType
	descriptionEntity["project"] = project
	descriptionEntity["domain"] = domain
	descriptionEntity["name"] = name
	descriptionEntity["version"] = version
	descriptionEntity["Digest"] = digest
	descriptionEntity["ShortDescription"] = shortDescription
	return descriptionEntity
}
