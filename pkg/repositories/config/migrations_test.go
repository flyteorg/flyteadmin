package config

import (
	"context"
	"fmt"
	"github.com/flyteorg/flytestdlib/database"
	"github.com/go-gormigrate/gormigrate/v2"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
	"time"

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

func TestNoopMigrations(t *testing.T) {
	gLogger := gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})

	gormConfig := &gorm.Config{
		Logger:                                   gLogger,
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	var gormDb *gorm.DB
	pgConfig := database.PostgresConfig{
		Host:         "localhost",
		Port:         30001,
		DbName:       "migratecopy1",
		User:         "postgres",
		Password:     "postgres",
		ExtraOptions: "",
		Debug:        false,
	}
	ctx := context.Background()
	postgresDsn := database.PostgresDsn(ctx, pgConfig)
	dialector := postgres.Open(postgresDsn)
	gormDb, err := gorm.Open(dialector, gormConfig)
	assert.NoError(t, err)

	fmt.Println(gormDb)

	m := gormigrate.New(gormDb, gormigrate.DefaultOptions, Migrations)
	if err := m.Migrate(); err != nil {
		fmt.Errorf("database migration failed: %v", err)
	}
	fmt.Println(ctx, "Migration ran successfully")
}

// Before running, create database replicator;
func TestMigrationReplicate(t *testing.T) {
	gLogger := gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})

	gormConfig := &gorm.Config{
		Logger:                                   gLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	var gormDb *gorm.DB
	pgConfig := database.PostgresConfig{
		Host:         "localhost",
		Port:         30001,
		DbName:       "replicator2", // this should be a completely blank database
		User:         "postgres",
		Password:     "postgres",
		ExtraOptions: "",
		Debug:        false,
	}
	ctx := context.Background()
	postgresDsn := database.PostgresDsn(ctx, pgConfig)
	dialector := postgres.Open(postgresDsn)
	gormDb, err := gorm.Open(dialector, gormConfig)
	assert.NoError(t, err)

	fmt.Println(gormDb)

	m := gormigrate.New(gormDb, gormigrate.DefaultOptions, NoopMigrations)
	if err := m.Migrate(); err != nil {
		fmt.Errorf("database migration failed: %v", err)
	}
	fmt.Println(ctx, "Migration ran successfully")
}

func TestLegacyOnlyMigrations(t *testing.T) {
	gLogger := gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})

	gormConfig := &gorm.Config{
		Logger:                                   gLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	var gormDb *gorm.DB
	pgConfig := database.PostgresConfig{
		Host:         "localhost",
		Port:         30001,
		DbName:       "currentempty",
		User:         "postgres",
		Password:     "postgres",
		ExtraOptions: "",
		Debug:        false,
	}
	ctx := context.Background()
	postgresDsn := database.PostgresDsn(ctx, pgConfig)
	dialector := postgres.Open(postgresDsn)
	gormDb, err := gorm.Open(dialector, gormConfig)
	assert.NoError(t, err)

	fmt.Println(gormDb)

	m := gormigrate.New(gormDb, gormigrate.DefaultOptions, LegacyMigrations)
	if err := m.Migrate(); err != nil {
		fmt.Errorf("database migration failed: %v", err)
	}
	fmt.Println(ctx, "Migration ran successfully")
}
