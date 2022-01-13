package entrypoints

import (
	"context"

	"github.com/flyteorg/flytestdlib/promutils"

	util "github.com/flyteorg/flyteadmin/cmd/entrypoints/util"
	"github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/flyteorg/flyteadmin/pkg/clusterresource"
	executioncluster "github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"

	repositoryConfig "github.com/flyteorg/flyteadmin/pkg/repositories/config"
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

func getClusterResourceController(ctx context.Context) clusterresource.Controller {
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

	return clusterresource.NewClusterResourceController(clientSet.AdminClient(), listTargetsProvider, scope)
}

var controllerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start a cluster resource controller to periodically sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		clusterResourceController := getClusterResourceController(ctx)
		clusterResourceController.Run()
		logger.Infof(ctx, "ClusterResourceController started running successfully")
	},
}

var controllerSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "This command will sync cluster resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		clusterResourceController := getClusterResourceController(ctx)
		err := clusterResourceController.Sync(ctx)
		if err != nil {
			logger.Fatalf(ctx, "Failed to sync cluster resources [%+v]", err)
		}
		logger.Infof(ctx, "ClusterResourceController synced successfully")
	},
}

func init() {
	RootCmd.AddCommand(parentClusterResourceCmd)
	parentClusterResourceCmd.AddCommand(controllerRunCmd)
	parentClusterResourceCmd.AddCommand(controllerSyncCmd)
}
