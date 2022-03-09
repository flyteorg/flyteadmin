package repositories

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	repoErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/jackc/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const pqInvalidDBCode = "3D000"
const defaultDB = "postgres"

// getGormLogLevel converts between the flytestdlib configured log level to the equivalent gorm log level.
func getGormLogLevel(ctx context.Context, logConfig *logger.Config) gormLogger.LogLevel {
	if logConfig == nil {
		logger.Debugf(ctx, "No log config block found, setting gorm db log level to: error")
		return gormLogger.Error
	}
	switch logConfig.Level {
	case logger.PanicLevel:
		fallthrough
	case logger.FatalLevel:
		fallthrough
	case logger.ErrorLevel:
		return gormLogger.Error
	case logger.WarnLevel:
		return gormLogger.Warn
	case logger.InfoLevel:
		fallthrough
	case logger.DebugLevel:
		return gormLogger.Info
	default:
		return gormLogger.Silent
	}
}

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
func getPostgresDsn(ctx context.Context, pgConfig runtimeInterfaces.PostgresConfig) string {
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
	gormDb *gorm.DB, err error) {
	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}
	gormConfig := &gorm.Config{
		Logger:                                   gormLogger.Default.LogMode(getGormLogLevel(ctx, logConfig)),
		DisableForeignKeyConstraintWhenMigrating: !dbConfig.EnableForeignKeyConstraintWhenMigrating,
	}

	// TODO: add other gorm-supported db type handling in further case blocks.
	switch {
	// TODO: Figure out a better proxy for a non-empty postgres config
	case len(dbConfig.PostgresConfig.Host) > 0 || len(dbConfig.PostgresConfig.User) > 0 || len(dbConfig.PostgresConfig.DbName) > 0:
		gormDb, err = gorm.Open(postgres.Open(getPostgresDsn(ctx, dbConfig.PostgresConfig)), gormConfig)
		if err != nil {
			gormDb, err = createPostgresDbIfNotExists(ctx, gormConfig, dbConfig.PostgresConfig)
		}
	case len(dbConfig.DeprecatedHost) > 0 || len(dbConfig.DeprecatedUser) > 0 || len(dbConfig.DeprecatedDbName) > 0:
		pgConfig := runtimeInterfaces.PostgresConfig{
			Host:         dbConfig.DeprecatedHost,
			Port:         dbConfig.DeprecatedPort,
			DbName:       dbConfig.DeprecatedDbName,
			User:         dbConfig.DeprecatedUser,
			Password:     dbConfig.DeprecatedPassword,
			PasswordPath: dbConfig.DeprecatedPasswordPath,
			ExtraOptions: dbConfig.DeprecatedExtraOptions,
			Debug:        dbConfig.DeprecatedDebug,
		}
		gormDb, err = gorm.Open(postgres.Open(getPostgresDsn(ctx, pgConfig)), gormConfig)
		if err != nil {
			gormDb, err = createPostgresDbIfNotExists(ctx, gormConfig, dbConfig.PostgresConfig)
		}
	default:
		panic(fmt.Sprintf("Unrecognized database config %v", dbConfig))
	}

	// Setup connection pool settings
	return gormDb, setupDbConnectionPool(gormDb, dbConfig)
}

// Creates DB if it doesn't exist for the passed in config
func createPostgresDbIfNotExists(ctx context.Context, gormConfig *gorm.Config, pgConfig runtimeInterfaces.PostgresConfig) (*gorm.DB, error) {

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
	defaultDBDialector := postgres.Open(getPostgresDsn(ctx, pgConfig))
	gormDb, err = gorm.Open(defaultDBDialector, gormConfig)
	if err != nil {
		return nil, err
	}

	type DatabaseResult struct {
		Exists bool
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

func setupDbConnectionPool(gormDb *gorm.DB, dbConfig *runtimeInterfaces.DbConfig) error {
	genericDb, err := gormDb.DB()
	if err != nil {
		return err
	}
	genericDb.SetConnMaxLifetime(dbConfig.ConnMaxLifeTime.Duration)
	genericDb.SetMaxIdleConns(dbConfig.MaxIdleConnections)
	genericDb.SetMaxOpenConns(dbConfig.MaxOpenConnections)
	return nil
}
