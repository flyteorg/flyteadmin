package entrypoints

import (
	"context"

	util "github.com/flyteorg/flyteadmin/cmd/entrypoints/util"
	"github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/flyteorg/flyteadmin/pkg/clusterresource"
	executioncluster "github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"

	repositoryConfig "github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

// TokenCacheKeyringProvider wraps the logic to save and retrieve tokens from the OS's keyring implementation.
type TokenCacheKeyringProvider struct {
	ServiceName string
	ServiceUser string
}

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
		initializationErrorCounter := scope.MustNewCounter(
			"flyteclient_initialization_error",
			"count of errors encountered initializing a flyte client from kube config")
		listTargetsProvider, err := executioncluster.NewListTargets(initializationErrorCounter, executioncluster.NewExecutionTargetProvider(), configuration.ClusterConfiguration())
		if err != nil {
			panic(err)
		}

		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).
			WithTokenCache(util.TokenCacheKeyringProvider{
				ServiceUser: util.KeyRingServiceUser,
				ServiceName: util.KeyRingServiceName,
			}).Build(ctx)
		if err != nil {
			panic(err)
		}

		clusterResourceController := clusterresource.NewClusterResourceController(clientSet.AdminClient(), listTargetsProvider, scope)
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
		initializationErrorCounter := scope.MustNewCounter(
			"flyteclient_initialization_error",
			"count of errors encountered initializing a flyte client from kube config")
		listTargetsProvider, err := executioncluster.NewListTargets(initializationErrorCounter, executioncluster.NewExecutionTargetProvider(), configuration.ClusterConfiguration())
		if err != nil {
			panic(err)
		}

		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).
			WithTokenCache(util.TokenCacheKeyringProvider{
				ServiceUser: util.KeyRingServiceUser,
				ServiceName: util.KeyRingServiceName,
			}).Build(ctx)
		if err != nil {
			panic(err)
		}

		clusterResourceController := clusterresource.NewClusterResourceController(clientSet.AdminClient(), listTargetsProvider, scope)
		err = clusterResourceController.Sync(ctx)
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
