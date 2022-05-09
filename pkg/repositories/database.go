package repositories

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/flyteorg/flytestdlib/database"

	repoErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"gorm.io/driver/sqlite"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/jackc/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const pqInvalidDBCode = "3D000"
const defaultDB = "postgres"

// Resolves a password value from either a user-provided inline value or a filepath whose contents contain a password.
func resolvePassword(ctx context.Context, passwordVal, passwordPath string) string {
	password := passwordVal
	if len(passwordPath) > 0 {
		if _, err := os.Stat(passwordPath); os.IsNotExist(err) {
			logger.Fatalf(ctx,
				"missing database password at specified path [%s]", passwordPath)
		}
		passwordVal, err := ioutil.ReadFile(passwordPath)
		if err != nil {
			logger.Fatalf(ctx, "failed to read database password from path [%s] with err: %v",
				passwordPath, err)
		}
		// Passwords can contain special characters as long as they are percent encoded
		// https://www.postgresql.org/docs/current/libpq-connect.html
		password = strings.TrimSpace(string(passwordVal))
	}
	return password
}

// Produces the DSN (data source name) for opening a postgres db connection.
func getPostgresDsn(ctx context.Context, pgConfig *runtimeInterfaces.PostgresConfig) string {
	password := resolvePassword(ctx, pgConfig.Password, pgConfig.PasswordPath)
	if len(password) == 0 {
		// The password-less case is included for development environments.
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User, password, pgConfig.ExtraOptions)
}

// GetDB uses the dbConfig to create gorm DB object. If the db doesn't exist for the dbConfig then a new one is created
// using the default db for the provider. eg : postgres has default dbName as postgres
func GetDB(ctx context.Context, dbConfig *runtimeInterfaces.DbConfig, logConfig *logger.Config) (
	*gorm.DB, error) {
	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}
	gormConfig := &gorm.Config{
		Logger:                                   database.GetGormLogger(ctx, logConfig),
		DisableForeignKeyConstraintWhenMigrating: !dbConfig.EnableForeignKeyConstraintWhenMigrating,
	}

	var gormDb *gorm.DB
	var err error

	switch {
	case dbConfig.SQLiteConfig != nil:
		if dbConfig.SQLiteConfig.File == "" {
			return nil, fmt.Errorf("illegal sqlite database configuration. `file` is a required parameter and should be a path")
		}
		gormDb, err = gorm.Open(sqlite.Open(dbConfig.SQLiteConfig.File), gormConfig)
		if err != nil {
			return nil, err
		}
	case dbConfig.PostgresConfig != nil && (len(dbConfig.PostgresConfig.Host) > 0 || len(dbConfig.PostgresConfig.User) > 0 || len(dbConfig.PostgresConfig.DbName) > 0):
		gormDb, err = createPostgresDbIfNotExists(ctx, gormConfig, dbConfig.PostgresConfig)
		if err != nil {
			return nil, err
		}

	case len(dbConfig.DeprecatedHost) > 0 || len(dbConfig.DeprecatedUser) > 0 || len(dbConfig.DeprecatedDbName) > 0:
		pgConfig := &runtimeInterfaces.PostgresConfig{
			Host:         dbConfig.DeprecatedHost,
			Port:         dbConfig.DeprecatedPort,
			DbName:       dbConfig.DeprecatedDbName,
			User:         dbConfig.DeprecatedUser,
			Password:     dbConfig.DeprecatedPassword,
			PasswordPath: dbConfig.DeprecatedPasswordPath,
			ExtraOptions: dbConfig.DeprecatedExtraOptions,
			Debug:        dbConfig.DeprecatedDebug,
		}
		gormDb, err = createPostgresDbIfNotExists(ctx, gormConfig, pgConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unrecognized database config, %v. Supported only postgres and sqlite", dbConfig)
	}

	// Setup connection pool settings
	return gormDb, setupDbConnectionPool(gormDb, dbConfig)
}

// Creates DB if it doesn't exist for the passed in config
func createPostgresDbIfNotExists(ctx context.Context, gormConfig *gorm.Config, pgConfig *runtimeInterfaces.PostgresConfig) (*gorm.DB, error) {

	dialector := postgres.Open(getPostgresDsn(ctx, pgConfig))
	gormDb, err := gorm.Open(dialector, gormConfig)
	if err == nil {
		return gormDb, nil
	}

	// Check if its invalid db code error
	cErr, ok := err.(repoErrors.ConnectError)
	if !ok {
		logger.Errorf(ctx, "Failed to cast error of type: %v, err: %v", reflect.TypeOf(err),
			err)
		return nil, err
	}
	pqError := cErr.Unwrap().(*pgconn.PgError)
	if pqError.Code != pqInvalidDBCode {
		return nil, err
	}

	logger.Warningf(ctx, "Database [%v] does not exist", pgConfig.DbName)

	// Every postgres installation includes a 'postgres' database by default. We connect to that now in order to
	// initialize the user-specified database.
	defaultDbPgConfig := pgConfig
	defaultDbPgConfig.DbName = defaultDB
	defaultDBDialector := postgres.Open(getPostgresDsn(ctx, defaultDbPgConfig))
	gormDb, err = gorm.Open(defaultDBDialector, gormConfig)
	if err != nil {
		return nil, err
	}

	// Because we asserted earlier that the db does not exist, we create it now.
	logger.Infof(ctx, "Creating database %v", pgConfig.DbName)

	// NOTE: golang sql drivers do not support parameter injection for CREATE calls
	createDBStatement := fmt.Sprintf("CREATE DATABASE %s", pgConfig.DbName)
	result := gormDb.Exec(createDBStatement)

	if result.Error != nil {
		return nil, result.Error
	}
	// Now try connecting to the db again
	return gorm.Open(dialector, gormConfig)
}

func setupDbConnectionPool(ctx context.Context, gormDb *gorm.DB, dbConfig *runtimeInterfaces.DbConfig) error {
	genericDb, err := gormDb.DB()
	if err != nil {
		return err
	}
	genericDb.SetConnMaxLifetime(dbConfig.ConnMaxLifeTime.Duration)
	genericDb.SetMaxIdleConns(dbConfig.MaxIdleConnections)
	genericDb.SetMaxOpenConns(dbConfig.MaxOpenConnections)
	logger.Infof(ctx, "Set connection pool values to [%+v]", genericDb.Stats())
	return nil
}
