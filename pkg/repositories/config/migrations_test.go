package config

import (
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAlterTableColumnType(t *testing.T) {
	gormDb := GetDbForTest(t)
	db, err := gormDb.DB()
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	query := GlobalMock.NewMock()
	query.WithQuery(
		`ALTER TABLE IF EXISTS execution_events ALTER COLUMN "id" TYPE bigint`)
	assert.NoError(t, err)
	tables = []string{"execution_events"}
	_ = alterTableColumnType(db, "id", "bigint")
	assert.True(t, query.Triggered)
}

func GetDbForTest(t *testing.T) *gorm.DB {
	mocket.Catcher.Register()
	db, err := gorm.Open(postgres.New(postgres.Config{DriverName: mocket.DriverName}))
	if err != nil {
		t.Fatalf("Failed to open mock db with err %v", err)
	}
	return db
}

func TestShouldApplyFixParentidMigration(t *testing.T) {
	gormDb := GetDbForTest(t)
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	query := GlobalMock.NewMock()
	query.WithQuery(`
	SELECT data_type
	FROM information_schema.columns
	WHERE table_name = $1 AND column_name = $2;
	`)

	shouldApply, err := shouldApplyFixParentidMigration(gormDb)
	assert.True(t, shouldApply)
	assert.True(t, query.Triggered)
	assert.NoError(t, err)
}
