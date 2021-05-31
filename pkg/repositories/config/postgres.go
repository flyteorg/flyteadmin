package config

import (
	"fmt"
	"time"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Required to import database driver.
)

const (
	Postgres                     = "postgres"
	DefaultMaxIdleConnection     = 25
	DefaultMaxOpenConnection     = 1
	DefaultConnectionMaxLifetime = time.Duration(1000)
)

// Generic interface for providing a config necessary to open a database connection.
type DbConnectionConfigProvider interface {
	// Returns the database type. For instance PostgreSQL or MySQL.
	GetType() string
	// Returns arguments specific for the database type necessary to open a database connection.
	GetArgs() string
	// Returns arguments specific for the database type necessary to set MaxOpenConnection.
	GetMaxOpenConnection() int
	// Returns arguments specific for the database type necessary to set MaxIdleConnection.
	GetMaxIdleConnection() int
	// Returns arguments specific for the database type necessary to set ConnectionMaxLifetime.
	GetConnectionMaxLifetime() time.Duration
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

func (p *PostgresConfigProvider) GetMaxOpenConnection() int {
	if p.config.MaxOpenConnection > 0 {
		return p.config.MaxOpenConnection
	}
	return DefaultMaxOpenConnection
}

func (p *PostgresConfigProvider) GetMaxIdleConnection() int {
	if p.config.MaxIdleConnection > 0 {
		return p.config.MaxIdleConnection
	}
	return DefaultMaxIdleConnection
}

func (p *PostgresConfigProvider) GetConnectionMaxLifetime() time.Duration {
	if p.config.ConnectionMaxLifetime > time.Duration(0) {
		return p.config.ConnectionMaxLifetime
	}
	return DefaultConnectionMaxLifetime
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
func OpenDbConnection(config DbConnectionConfigProvider) *gorm.DB {
	db, err := gorm.Open(config.GetType(), config.GetArgs())
	if err != nil {
		panic(err)
	}
	db.LogMode(config.IsDebug())
	db.DB().SetMaxOpenConns(config.GetMaxOpenConnection())
	db.DB().SetMaxIdleConns(config.GetMaxIdleConnection())
	db.DB().SetConnMaxLifetime(config.GetConnectionMaxLifetime())
	return db
}
