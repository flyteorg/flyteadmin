package entrypoints

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/flyteorg/flyteadmin/pkg/common"
	repositoryCommonConfig "github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	scheduler "github.com/flyteorg/flyteadmin/scheduler/executor"
	schdulerRepoConfig "github.com/flyteorg/flyteadmin/scheduler/repositories"
	"github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	_ "github.com/jinzhu/gorm/dialects/postgres" // Required to import database driver.
	"github.com/spf13/cobra"
)

var schedulerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will start the flyte native scheduler and periodically get new schedules from the db for scheduling",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()

		// Define the schedulerScope for prometheus metrics
		schedulerScope := promutils.NewScope(applicationConfiguration.MetricsScope).NewSubScope("flytescheduler")

		defer func() {
			if err := recover(); err != nil {
				schedulerScope.MustNewCounter("initialization_panic",
					"panics encountered initializing the native flyte scheduler").Inc()
				logger.Fatalf(context.Background(), fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
			}
		}()

		dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
		dbConfig := repositoryCommonConfig.DbConfig{
			BaseConfig: repositoryCommonConfig.BaseConfig{
				IsDebug: dbConfigValues.Debug,
			},
			Host:         dbConfigValues.Host,
			Port:         dbConfigValues.Port,
			DbName:       dbConfigValues.DbName,
			User:         dbConfigValues.User,
			Password:     dbConfigValues.Password,
			ExtraOptions: dbConfigValues.ExtraOptions,
		}
		db := schdulerRepoConfig.GetRepository(
			schdulerRepoConfig.POSTGRES, dbConfig, schedulerScope.NewSubScope("database"))

		clientSet, err := admin.ClientSetBuilder().WithConfig(admin.GetConfig(ctx)).Build(ctx)
		if err != nil {
			panic(err)
		}
		adminServiceClient := clientSet.AdminClient()

		schedulerWorkflowExecutor := scheduler.NewWorkflowExecutor(db, configuration, schedulerScope, adminServiceClient)

		logger.Info(context.Background(), "Successfully initialized a native flyte scheduler")

		schedulerWorkflowExecutor.Run()

		logger.Infof(ctx, "Flyte native scheduler started successfully")
	},
}

func init() {
	RootCmd.AddCommand(schedulerRunCmd)

	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey,
		contextutils.ExecIDKey, contextutils.WorkflowIDKey, contextutils.NodeIDKey, contextutils.TaskIDKey,
		contextutils.TaskTypeKey, common.RuntimeTypeKey, common.RuntimeVersionKey)
}
