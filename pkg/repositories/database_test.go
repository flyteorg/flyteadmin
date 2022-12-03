package repositories

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/database"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetupDbConnectionPool(t *testing.T) {
	ctx := context.TODO()
	t.Run("successful", func(t *testing.T) {
		gormDb, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), &gorm.Config{})
		assert.Nil(t, err)
		dbConfig := &database.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
			ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
		}
		err = setupDbConnectionPool(ctx, gormDb, dbConfig)
		assert.Nil(t, err)
		genericDb, err := gormDb.DB()
		assert.Nil(t, err)
		assert.Equal(t, genericDb.Stats().MaxOpenConnections, 1000)
	})
	t.Run("unset duration", func(t *testing.T) {
		gormDb, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), &gorm.Config{})
		assert.Nil(t, err)
		dbConfig := &database.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
		}
		err = setupDbConnectionPool(ctx, gormDb, dbConfig)
		assert.Nil(t, err)
		genericDb, err := gormDb.DB()
		assert.Nil(t, err)
		assert.Equal(t, genericDb.Stats().MaxOpenConnections, 1000)
	})
	t.Run("failed to get DB", func(t *testing.T) {
		gormDb := &gorm.DB{
			Config: &gorm.Config{
				ConnPool: &gorm.PreparedStmtDB{},
			},
		}
		dbConfig := &database.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
			ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
		}
		err := setupDbConnectionPool(ctx, gormDb, dbConfig)
		assert.NotNil(t, err)
	})
}

func TestGetDB(t *testing.T) {
	ctx := context.TODO()

	t.Run("missing DB Config", func(t *testing.T) {
		_, err := GetDB(ctx, &database.DbConfig{}, &logger.Config{})
		assert.Error(t, err)
	})

	t.Run("sqlite config", func(t *testing.T) {
		dbFile := path.Join(t.TempDir(), "admin.db")
		db, err := GetDB(ctx, &database.DbConfig{
			SQLite: database.SQLiteConfig{
				File: dbFile,
			},
		}, &logger.Config{})
		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.FileExists(t, dbFile)
		assert.Equal(t, "sqlite", db.Name())
	})
}
