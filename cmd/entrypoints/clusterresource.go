package entrypoints

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/clusterresource"
	executioncluster "github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"

	"github.com/flyteorg/flyteadmin/pkg/runtime"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repositoryConfig "github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

var parentClusterResourceCmd = &cobra.Command{
	Use:   "clusterresource",
	Short: "This command administers the ClusterResourceController. Please choose a subcommand.",
}

func GetLocalDbConfig() repositoryConfig.DbConfig {
	return repositoryConfig.DbConfig{
		Host:   "localhost",
		Port:   5432,
		DbName: "postgres",
		User:   "postgres",
	}
}

var controllerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start a cluster resource controller to periodically sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
		dbConfig := repositoryConfig.DbConfig{
			BaseConfig: repositoryConfig.BaseConfig{
				LogLevel:                                 dbConfigValues.LogLevel,
				DisableForeignKeyConstraintWhenMigrating: dbConfigValues.DisableForeignKeyConstraintWhenMigrating,
			},
			Host:         dbConfigValues.Host,
			Port:         dbConfigValues.Port,
			DbName:       dbConfigValues.DbName,
			User:         dbConfigValues.User,
			Password:     dbConfigValues.Password,
			ExtraOptions: dbConfigValues.ExtraOptions,
		}
		db := repositories.GetRepository(
			repositories.POSTGRES, dbConfig, scope.NewSubScope("database"))

		cfg := config.GetConfig()
		executionCluster := executioncluster.GetExecutionCluster(
			scope.NewSubScope("cluster"),
			cfg.KubeConfig,
			cfg.Master,
			configuration,
			db)

		clusterResourceController := clusterresource.NewClusterResourceController(db, executionCluster, scope)
		clusterResourceController.Run()
		logger.Infof(ctx, "ClusterResourceController started successfully")
	},
}

var controllerSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "This command will sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().MetricsScope).NewSubScope("clusterresource")
		dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
		dbConfig := repositoryConfig.DbConfig{
			BaseConfig: repositoryConfig.BaseConfig{
				LogLevel:                                 dbConfigValues.LogLevel,
				DisableForeignKeyConstraintWhenMigrating: dbConfigValues.DisableForeignKeyConstraintWhenMigrating,
			},
			Host:         dbConfigValues.Host,
			Port:         dbConfigValues.Port,
			DbName:       dbConfigValues.DbName,
			User:         dbConfigValues.User,
			Password:     dbConfigValues.Password,
			ExtraOptions: dbConfigValues.ExtraOptions,
		}
		db := repositories.GetRepository(
			repositories.POSTGRES, dbConfig, scope.NewSubScope("database"))

		cfg := config.GetConfig()
		executionCluster := executioncluster.GetExecutionCluster(
			scope.NewSubScope("cluster"),
			cfg.KubeConfig,
			cfg.Master,
			configuration,
			db)

		clusterResourceController := clusterresource.NewClusterResourceController(db, executionCluster, scope)
		err := clusterResourceController.Sync(ctx)
		if err != nil {
			logger.Fatalf(ctx, "Failed to sync cluster resources [%+v]", err)
		}
		logger.Infof(ctx, "ClusterResourceController started successfully")
	},
}

func init() {
	RootCmd.AddCommand(parentClusterResourceCmd)
	parentClusterResourceCmd.AddCommand(controllerRunCmd)
	parentClusterResourceCmd.AddCommand(controllerSyncCmd)
}
