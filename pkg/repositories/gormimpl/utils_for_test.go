// Shared utils for postgresql tests.
package gormimpl

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/common"

	mocket "github.com/Selvatico/go-mocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const project = "project"
const domain = "domain"
const name = "name"
const description = "description"
const resourceType = core.ResourceType_WORKFLOW
const version = "XYZ"

func GetDbForTest(t *testing.T) *gorm.DB {
	mocket.Catcher.Register()
	db, err := gorm.Open(postgres.New(postgres.Config{DriverName: mocket.DriverName}))
	if err != nil {
		t.Fatalf("Failed to open mock db with err %v", err)
	}
	return db
}

func getEqualityFilter(entity common.Entity, field string, value interface{}) common.InlineFilter {
	filter, _ := common.NewSingleValueFilter(entity, common.Equal, field, value)
	return filter
}

func makeSortParameters(t *testing.T, direction admin.Sort_Direction, key string) []common.SortParameter {
	params, err := common.NewSortParameter(&admin.Sort{
		Direction: direction,
		Key:       key,
	}, sets.NewString(key))
	require.NoError(t, err)
	return params
}
