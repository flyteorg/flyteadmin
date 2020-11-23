package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateProject(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	query.WithQuery(
		`INSERT INTO "projects" ("created_at","updated_at","deleted_at","identifier","name","description","labels","state") VALUES (?,?,?,?,?,?,?,?)`)

	activeState := int32(admin.Project_ACTIVE)
	err := projectRepo.Create(context.Background(), models.Project{
		Identifier:  "proj",
		Name:        "proj",
		Description: "projDescription",
		State:       &activeState,
	})
	assert.NoError(t, err)
	assert.True(t, query.Triggered)
}

func TestGetProject(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response["identifier"] = "project_id"
	response["name"] = "project_name"
	response["description"] = "project_description"
	response["state"] = admin.Project_ACTIVE

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "projects"  WHERE "projects"."deleted_at" IS NULL AND ` +
		`(("projects"."identifier" = project_id)) ORDER BY "projects"."identifier" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := projectRepo.Get(context.Background(), "project_id")
	assert.Nil(t, err)
	assert.Equal(t, "project_id", output.Identifier)
	assert.Equal(t, "project_name", output.Name)
	assert.Equal(t, "project_description", output.Description)
	assert.Equal(t, int32(admin.Project_ACTIVE), *output.State)
}

func TestListAllProjects(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	projects := make([]map[string]interface{}, 2)
	fooProject := make(map[string]interface{})
	fooProject["identifier"] = "foo"
	fooProject["name"] = "foo =)"
	fooProject["description"] = "foo description"
	fooProject["state"] = admin.Project_ACTIVE
	projects[0] = fooProject

	barProject := make(map[string]interface{})
	barProject["identifier"] = "bar"
	barProject["name"] = "Bar"
	barProject["description"] = "Bar description"
	barProject["state"] = admin.Project_ACTIVE
	projects[1] = barProject

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "projects"  WHERE "projects"."deleted_at" IS NULL AND ((state != 1)) ORDER BY identifier asc`).
		WithReply(projects)

	var alphabeticalSortParam, _ = common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "identifier",
	})
	output, err := projectRepo.ListAll(context.Background(), alphabeticalSortParam)
	assert.Nil(t, err)
	assert.Len(t, output, 2)
	assert.Equal(t, "foo", output[0].Identifier)
	assert.Equal(t, "foo =)", output[0].Name)
	assert.Equal(t, "foo description", output[0].Description)
	assert.Equal(t, int32(admin.Project_ACTIVE), *output[0].State)
	assert.Equal(t, "bar", output[1].Identifier)
	assert.Equal(t, "Bar", output[1].Name)
	assert.Equal(t, "Bar description", output[1].Description)
	assert.Equal(t, int32(admin.Project_ACTIVE), *output[1].State)
}

func TestListProjects(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	projects := make([]map[string]interface{}, 1)
	fooProject := make(map[string]interface{})
	fooProject["identifier"] = "foo"
	fooProject["name"] = "foo =)"
	fooProject["description"] = "foo description"
	fooProject["state"] = admin.Project_ACTIVE
	projects[0] = fooProject

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "projects"  WHERE "projects"."deleted_at" IS NULL AND ((name = foo)) ORDER BY identifier asc LIMIT 1 OFFSET 0`).
		WithReply(projects)

	var alphabeticalSortParam, _ = common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "identifier",
	})
	filter, err := common.NewSingleValueFilter(common.Project, common.Equal, "name", "foo")
	assert.Nil(t, err)
	output, err := projectRepo.List(context.Background(), interfaces.ListResourceInput{
		Offset:        0,
		Limit:         1,
		InlineFilters: []common.InlineFilter{filter},
		SortParameter: alphabeticalSortParam,
	})
	assert.Nil(t, err)
	assert.Len(t, output, 1)
	assert.Equal(t, "foo", output[0].Identifier)
	assert.Equal(t, "foo =)", output[0].Name)
	assert.Equal(t, "foo description", output[0].Description)
	assert.Equal(t, int32(admin.Project_ACTIVE), *output[0].State)
}

func TestListProjects_NoFilters(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	projects := make([]map[string]interface{}, 1)
	fooProject := make(map[string]interface{})
	fooProject["identifier"] = "foo"
	fooProject["name"] = "foo =)"
	fooProject["description"] = "foo description"
	fooProject["state"] = admin.Project_ACTIVE
	projects[0] = fooProject

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "projects"  WHERE "projects"."deleted_at" IS NULL AND ((state != 1)) ORDER BY identifier asc LIMIT 1 OFFSET 0`).
		WithReply(projects)

	var alphabeticalSortParam, _ = common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "identifier",
	})
	output, err := projectRepo.List(context.Background(), interfaces.ListResourceInput{
		Offset:        0,
		Limit:         1,
		SortParameter: alphabeticalSortParam,
	})
	assert.Nil(t, err)
	assert.Len(t, output, 1)
	assert.Equal(t, "foo", output[0].Identifier)
	assert.Equal(t, "foo =)", output[0].Name)
	assert.Equal(t, "foo description", output[0].Description)
	assert.Equal(t, int32(admin.Project_ACTIVE), *output[0].State)
}

func TestListProjects_NoLimit(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	var alphabeticalSortParam, _ = common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "identifier",
	})
	_, err := projectRepo.List(context.Background(), interfaces.ListResourceInput{
		Limit:         0,
		SortParameter: alphabeticalSortParam,
	})
	assert.NotNil(t, err)
}

func TestUpdateProject(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	query.WithQuery(`UPDATE "projects" SET "description" = ?, "identifier" = ?, "name" = ?, "state" = ?, "updated_at" = ?  WHERE "projects"."deleted_at" IS NULL AND "projects"."identifier" = ?`)

	activeState := int32(admin.Project_ACTIVE)
	err := projectRepo.UpdateProject(context.Background(), models.Project{
		Identifier:  "project_id",
		Name:        "project_name",
		Description: "project_description",
		State:       &activeState,
	})
	assert.Nil(t, err)
	assert.True(t, query.Triggered)
}
