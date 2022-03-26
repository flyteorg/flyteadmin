package runtime

import (
	"time"

	datacatalogDbConfig "github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
)

const database = "database"
const flyteAdmin = "flyteadmin"
const scheduler = "scheduler"
const remoteData = "remoteData"
const notifications = "notifications"
const domains = "domains"
const externalEvents = "externalEvents"
const metricPort = 10254
const postgres = "postgres"

const KB = 1024
const MB = KB * KB

var databaseConfig, err = config.RegisterSection(database, &interfaces.DbConfig{
	DeprecatedPort:         5432,
	DeprecatedUser:         postgres,
	DeprecatedHost:         postgres,
	DeprecatedDbName:       postgres,
	DeprecatedExtraOptions: "sslmode=disable",
	MaxIdleConnections:     10,
	MaxOpenConnections:     1000,
	ConnMaxLifeTime:        config.Duration{Duration: time.Hour},
})
var flyteAdminConfig = config.MustRegisterSection(flyteAdmin, &interfaces.ApplicationConfig{
	ProfilerPort:          metricPort,
	MetricsScope:          "flyte:",
	MetadataStoragePrefix: []string{"metadata", "admin"},
	EventVersion:          2,
	AsyncEventsBufferSize: 100,
	MaxParallelism:        25,
	K8SServiceAccount:     "default",
})

var schedulerConfig = config.MustRegisterSection(scheduler, &interfaces.SchedulerConfig{
	ProfilerPort: config.Port{Port: metricPort},
	EventSchedulerConfig: interfaces.EventSchedulerConfig{
		Scheme:               common.Local,
		FlyteSchedulerConfig: &interfaces.FlyteSchedulerConfig{},
	},
	WorkflowExecutorConfig: interfaces.WorkflowExecutorConfig{
		Scheme: common.Local,
		FlyteWorkflowExecutorConfig: &interfaces.FlyteWorkflowExecutorConfig{
			AdminRateLimit: &interfaces.AdminRateLimit{
				Tps:   100,
				Burst: 10,
			},
		},
	},
})
var remoteDataConfig = config.MustRegisterSection(remoteData, &interfaces.RemoteDataConfig{
	Scheme:                common.None,
	MaxSizeInBytes:        2 * MB,
	InlineEventDataPolicy: interfaces.InlineEventDataPolicyOffload,
	SignedURL: interfaces.SignedURL{
		Enabled: false,
	},
})
var notificationsConfig = config.MustRegisterSection(notifications, &interfaces.NotificationsConfig{
	Type: common.Local,
})
var domainsConfig = config.MustRegisterSection(domains, &interfaces.DomainsConfig{
	{
		ID:   "development",
		Name: "development",
	},
	{
		ID:   "staging",
		Name: "staging",
	},
	{
		ID:   "production",
		Name: "production",
	},
})
var externalEventsConfig = config.MustRegisterSection(externalEvents, &interfaces.ExternalEventsConfig{
	Type: common.Local,
})

// Implementation of an interfaces.ApplicationConfiguration
type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetDbConfig() *interfaces.DbConfig {
	// The error only occurs when using single binary to launch flyte cluster.
	// The reason is that we try to register "database" section in flyteadmin and datacatalog
	// Here we will read database config from datacatalog if it exists.
	if err != nil {
		databaseSection := config.GetSection(database)
		datacatalogDatabaseConfig := databaseSection.GetConfig().(*datacatalogDbConfig.DbConfigSection)
		switch {
		case len(datacatalogDatabaseConfig.Host) > 0 || len(datacatalogDatabaseConfig.User) > 0 || len(datacatalogDatabaseConfig.DbName) > 0:
			return &interfaces.DbConfig{
				DeprecatedPort:     datacatalogDatabaseConfig.Port,
				DeprecatedUser:     datacatalogDatabaseConfig.User,
				DeprecatedHost:     datacatalogDatabaseConfig.Host,
				DeprecatedDbName:   datacatalogDatabaseConfig.DbName,
				MaxIdleConnections: 10,
				MaxOpenConnections: 1000,
				ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
			}
		case datacatalogDatabaseConfig.PostgresConfig != nil:
			return &interfaces.DbConfig{
				PostgresConfig: &interfaces.PostgresConfig{
					Host:         datacatalogDatabaseConfig.PostgresConfig.Host,
					Port:         datacatalogDatabaseConfig.PostgresConfig.Port,
					DbName:       datacatalogDatabaseConfig.PostgresConfig.DbName,
					User:         datacatalogDatabaseConfig.PostgresConfig.User,
					Password:     datacatalogDatabaseConfig.PostgresConfig.Password,
					PasswordPath: datacatalogDatabaseConfig.PostgresConfig.PasswordPath,
					ExtraOptions: datacatalogDatabaseConfig.PostgresConfig.ExtraOptions,
					Debug:        datacatalogDatabaseConfig.PostgresConfig.Debug,
				},
			}
		case datacatalogDatabaseConfig.SQLiteConfig != nil:
			return &interfaces.DbConfig{
				SQLiteConfig: &interfaces.SQLiteConfig{
					File: datacatalogDatabaseConfig.SQLiteConfig.File,
				},
			}
		default:
			return &interfaces.DbConfig{}
		}
	}

	return databaseConfig.GetConfig().(*interfaces.DbConfig)
}

func (p *ApplicationConfigurationProvider) GetTopLevelConfig() *interfaces.ApplicationConfig {
	return flyteAdminConfig.GetConfig().(*interfaces.ApplicationConfig)
}

func (p *ApplicationConfigurationProvider) GetSchedulerConfig() *interfaces.SchedulerConfig {
	return schedulerConfig.GetConfig().(*interfaces.SchedulerConfig)
}

func (p *ApplicationConfigurationProvider) GetRemoteDataConfig() *interfaces.RemoteDataConfig {
	return remoteDataConfig.GetConfig().(*interfaces.RemoteDataConfig)
}

func (p *ApplicationConfigurationProvider) GetNotificationsConfig() *interfaces.NotificationsConfig {
	return notificationsConfig.GetConfig().(*interfaces.NotificationsConfig)
}

func (p *ApplicationConfigurationProvider) GetDomainsConfig() *interfaces.DomainsConfig {
	return domainsConfig.GetConfig().(*interfaces.DomainsConfig)
}

func (p *ApplicationConfigurationProvider) GetExternalEventsConfig() *interfaces.ExternalEventsConfig {
	return externalEventsConfig.GetConfig().(*interfaces.ExternalEventsConfig)
}

func NewApplicationConfigurationProvider() interfaces.ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
