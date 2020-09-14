package runtime

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
)

const database = "database"
const flyteAdmin = "flyteadmin"
const scheduler = "scheduler"
const remoteData = "remoteData"
const notifications = "notifications"
const domains = "domains"

var databaseConfig = config.MustRegisterSection(database, &interfaces.DbConfigSection{})
var flyteAdminConfig = config.MustRegisterSection(flyteAdmin, &interfaces.ApplicationConfig{})
var schedulerConfig = config.MustRegisterSection(scheduler, &interfaces.SchedulerConfig{})
var remoteDataConfig = config.MustRegisterSection(remoteData, &interfaces.RemoteDataConfig{})
var notificationsConfig = config.MustRegisterSection(notifications, &interfaces.NotificationsConfig{})
var domainsConfig = config.MustRegisterSection(domains, &interfaces.DomainsConfig{})

// Implementation of an interfaces.ApplicationConfiguration
type ApplicationConfigurationProvider struct{}

func (p *ApplicationConfigurationProvider) GetDbConfig() interfaces.DbConfig {
	dbConfigSection := databaseConfig.GetConfig().(*interfaces.DbConfigSection)
	password := dbConfigSection.DBCertSection.Password
	if len(dbConfigSection.DBCertSection.PasswordPath) > 0 {
		if _, err := os.Stat(dbConfigSection.DBCertSection.PasswordPath); os.IsNotExist(err) {
			logger.Fatalf(context.Background(),
				"missing database password at specified path [%s]", dbConfigSection.DBCertSection.PasswordPath)
		}
		passwordVal, err := ioutil.ReadFile(dbConfigSection.DBCertSection.PasswordPath)
		if err != nil {
			logger.Fatalf(context.Background(), "failed to read database password from path [%s] with err: %v",
				dbConfigSection.DBCertSection.PasswordPath, err)
		}
		password = string(passwordVal)
	}
	rootCA := dbConfigSection.DBCertSection.RootCA
	if len(dbConfigSection.DBCertSection.RootCAPath) > 0 {
		if _, err := os.Stat(dbConfigSection.DBCertSection.RootCAPath); os.IsNotExist(err) {
			logger.Fatalf(context.Background(),
				"missing database root CA at specified path [%s]", dbConfigSection.DBCertSection.RootCAPath)
		}
		rootCAVal, err := ioutil.ReadFile(dbConfigSection.DBCertSection.RootCAPath)
		if err != nil {
			logger.Fatalf(context.Background(), "failed to read database root CA from path [%s] with err: %v",
				dbConfigSection.DBCertSection.RootCAPath, err)
		}
		rootCA = string(rootCAVal)
	}
	var region string
	if len(dbConfigSection.AWSDbConfig.Region) > 0 {
		region = dbConfigSection.AWSDbConfig.Region
	}

	return interfaces.DbConfig{
		Host:         dbConfigSection.Host,
		Port:         dbConfigSection.Port,
		DbName:       dbConfigSection.DbName,
		User:         dbConfigSection.User,
		Password:     password,
		RootCA:       rootCA,
		Region:       region,
		UseIAM:       dbConfigSection.AWSDbConfig.UseIAM,
		ExtraOptions: dbConfigSection.ExtraOptions,
	}
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
func NewApplicationConfigurationProvider() interfaces.ApplicationConfiguration {
	return &ApplicationConfigurationProvider{}
}
