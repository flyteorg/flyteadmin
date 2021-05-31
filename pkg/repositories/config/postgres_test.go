package config

import (
	"testing"
	"time"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

var (
	ExpectedMaxIdleConnection     = 1
	ExpectedMaxOpenConnection     = 25
	ExpectedConnectionMaxLifetime = time.Duration(1000)
)

func TestConstructGormArgs(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		BaseConfig: BaseConfig{
			IsDebug: true,
		},
		Host:                  "localhost",
		Port:                  5432,
		DbName:                "postgres",
		User:                  "postgres",
		ExtraOptions:          "sslmode=disable",
		MaxIdleConnection:     ExpectedMaxIdleConnection,
		MaxOpenConnection:     ExpectedMaxOpenConnection,
		ConnectionMaxLifetime: ExpectedConnectionMaxLifetime,
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres sslmode=disable", postgresConfigProvider.GetArgs())
	assert.True(t, postgresConfigProvider.IsDebug())
	assert.Equal(t, ExpectedMaxOpenConnection, postgresConfigProvider.GetMaxOpenConnection())
	assert.Equal(t, ExpectedConnectionMaxLifetime, postgresConfigProvider.GetConnectionMaxLifetime())
	assert.Equal(t, ExpectedMaxIdleConnection, postgresConfigProvider.GetMaxIdleConnection())
}

func TestDefaultConfigArgs(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		BaseConfig: BaseConfig{
			IsDebug: true,
		},
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		ExtraOptions: "sslmode=disable",
	}, mockScope.NewTestScope())

	assert.Equal(t, DefaultMaxOpenConnection, postgresConfigProvider.GetMaxOpenConnection())
	assert.Equal(t, DefaultConnectionMaxLifetime, postgresConfigProvider.GetConnectionMaxLifetime())
	assert.Equal(t, DefaultMaxIdleConnection, postgresConfigProvider.GetMaxIdleConnection())
}

func TestConstructGormArgsWithPassword(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		Password:     "pass",
		ExtraOptions: "sslmode=enable",
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass sslmode=enable", postgresConfigProvider.GetArgs())
}

func TestConstructGormArgsWithPasswordNoExtra(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		Host:     "localhost",
		Port:     5432,
		DbName:   "postgres",
		User:     "postgres",
		Password: "pass",
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass ", postgresConfigProvider.GetArgs())
}
