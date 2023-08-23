package gormimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

func Test_modelColumns(t *testing.T) {
	expected := sets.NewString(
		"closure",
		"created_at",
		"deleted_at",
		"domain",
		"duration",
		"execution_domain",
		"execution_name",
		"execution_project",
		"id",
		"input_uri",
		"name",
		"node_id",
		"phase",
		"phase_version",
		"project",
		"retry_attempt",
		"started_at",
		"task_execution_created_at",
		"task_execution_updated_at",
		"updated_at",
		"version")

	actual := modelColumns(models.TaskExecution{})

	assert.Equal(t, expected, actual)
}
