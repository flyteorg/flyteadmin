package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"google.golang.org/grpc/codes"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Required to import database driver.
	// goGormPostgres "gorm.io/driver/postgres"
)

const Postgres = "postgres"

// Generic interface for providing a config necessary to open a database connection.
type DbConnectionConfigProvider interface {
	// Returns the database type. For instance PostgreSQL or MySQL.
	GetType() string
	// Returns arguments specific for the database type necessary to open a database connection.
	GetArgs() string
	// Enables verbose logging.
	WithDebugModeEnabled()
	// Disables verbose logging.
	WithDebugModeDisabled()
	// Returns whether verbose logging is enabled or not.
	IsDebug() bool
}

type BaseConfig struct {
	IsDebug bool
}

// PostgreSQL implementation for DbConnectionConfigProvider.
type PostgresConfigProvider struct {
	config DbConfig
	scope  promutils.Scope
}

// TODO : Make the Config provider itself env based
func NewPostgresConfigProvider(config DbConfig, scope promutils.Scope) DbConnectionConfigProvider {
	return &PostgresConfigProvider{
		config: config,
		scope:  scope,
	}
}

func (p *PostgresConfigProvider) GetType() string {
	return Postgres
}

func (p *PostgresConfigProvider) GetArgs() string {
	if p.config.Password == "" {
		// Switch for development
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			p.config.Host, p.config.Port, p.config.DbName, p.config.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		p.config.Host, p.config.Port, p.config.DbName, p.config.User, p.config.Password, p.config.ExtraOptions)
}

func (p *PostgresConfigProvider) WithDebugModeEnabled() {
	p.config.IsDebug = true
}

func (p *PostgresConfigProvider) WithDebugModeDisabled() {
	p.config.IsDebug = false
}

func (p *PostgresConfigProvider) IsDebug() bool {
	return p.config.IsDebug
}

// Opens a connection to the database specified in the config.
// You must call CloseDbConnection at the end of your session!
func OpenDbConnection(ctx context.Context, dbConfig DbConfig) (*gorm.DB, error) {
	connConfig := pgx.ConnConfig{
		Config: pgconn.Config{
			Host:     dbConfig.Host,
			User:     dbConfig.User,
			Database: dbConfig.DbName,
		},
	}

	if len(dbConfig.RootCA) > 0 {
		logger.Debugf(ctx, "Preparing to append a root CA")
		ca := x509.NewCertPool()
		if ok := ca.AppendCertsFromPEM([]byte(dbConfig.RootCA)); !ok {
			logger.Errorf(ctx, "Failed to append cert from PEM")
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Failed to append cert from PEM")
		}

		connConfig.TLSConfig = &tls.Config{
			RootCAs:    ca,
			ServerName: dbConfig.Host,
		}
	}

	// Use IAM auth when connecting to the database.
	if dbConfig.UseIAM {
		logger.Debugf(ctx, "Preparing to use IAM")
		sess, err := session.NewSession(
			&aws.Config{
				Region: aws.String(dbConfig.Region),
			},
		)
		if err != nil {
			logger.Errorf(ctx, "Failed to create new AWS session with err: [%+v]", err)
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Failed to create new AWS session with err: [%+v]", err)
		}

		endpoint := fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port)
		token, err := rdsutils.BuildAuthToken(
			endpoint,
			dbConfig.Region,
			dbConfig.User,
			sess.Config.Credentials,
		)
		if err != nil {
			logger.Errorf(ctx, "Failed to build auth token with err: [%+v]", err)
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Failed to build auth token with err: [%+v]", err)
		}

		connConfig.Password = token
	} else {
		logger.Debugf(ctx, "Not using IAM")
		connConfig.Password = dbConfig.Password
	}

	db, err := gorm.Open("postgres", connConfig.ConnString())
	if err != nil {
		logger.Errorf(ctx, "Failed to open db connection with err: [%+v]", err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Failed to open db connection with err: [%+v]", err)
	}
	db.LogMode(dbConfig.IsDebug)
	return db, nil
}
