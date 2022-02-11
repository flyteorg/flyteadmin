package repositories

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

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

func getPostgresDsn(ctx context.Context, pgConfig runtimeInterfaces.PostgresConfig) string {
	password := pgConfig.Password
	if len(pgConfig.PasswordPath) > 0 {
		if _, err := os.Stat(pgConfig.PasswordPath); os.IsNotExist(err) {
			logger.Fatalf(ctx,
				"missing database password at specified path [%s]", pgConfig.PasswordPath)
		}
		passwordVal, err := ioutil.ReadFile(pgConfig.PasswordPath)
		if err != nil {
			logger.Fatalf(ctx, "failed to read database password from path [%s] with err: %v",
				pgConfig.PasswordPath, err)
		}
		// Passwords can contain special characters as long as they are percent encoded
		// https://www.postgresql.org/docs/current/libpq-connect.html
		password = strings.TrimSpace(string(passwordVal))
	}

	if len(password) == 0 {
		// The password-less case is included for development environments.
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User, pgConfig.Password, pgConfig.ExtraOptions)
}

func GetDB(ctx context.Context, dbConfig *runtimeInterfaces.DbConfig, logConfig *logger.Config) (
	*gorm.DB, error) {
	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}
	var dialector gorm.Dialector
	logLevel := getGormLogLevel(ctx, logConfig)

	switch {
	case len(dbConfig.PostgresConfig.Host) > 0 || len(dbConfig.PostgresConfig.User) > 0 || len(dbConfig.PostgresConfig.DbName) > 0:

		dialector = postgres.Open(getPostgresDsn(ctx, dbConfig.PostgresConfig))
		// TODO: add other gorm-supported db type handling in further case blocks.
	default:
		panic(fmt.Sprintf("Unrecognized database config %v", dbConfig))
	}

	return gorm.Open(dialector, &gorm.Config{
		Logger:                                   gormLogger.Default.LogMode(logLevel),
		DisableForeignKeyConstraintWhenMigrating: dbConfig.DisableForeignKeyConstraintWhenMigrating,
	})
}
