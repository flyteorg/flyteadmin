package entrypoints

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"runtime/debug"

	eventWriter "github.com/flyteorg/flyteadmin/pkg/async/events/implementations"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications"
	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flyteadmin/pkg/data"
	executionCluster "github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"
	manager "github.com/flyteorg/flyteadmin/pkg/manager/impl"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repositoryConfig "github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	workflowengine "github.com/flyteorg/flyteadmin/pkg/workflowengine/impl"
	scheduler "github.com/flyteorg/flyteadmin/scheduler/executor"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"

	_ "github.com/jinzhu/gorm/dialects/postgres" // Required to import database driver.
	"github.com/spf13/cobra"
)


const defaultRetries = 3

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
					"panics encountered initializing the native flytescheduler").Inc()
				logger.Fatalf(context.Background(), fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
			}
		}()

		dbConfigValues := configuration.ApplicationConfiguration().GetDbConfig()
		dbConfig := repositoryConfig.DbConfig{
			BaseConfig: repositoryConfig.BaseConfig{
				IsDebug: dbConfigValues.Debug,
			},
			Host:         dbConfigValues.Host,
			Port:         dbConfigValues.Port,
			DbName:       dbConfigValues.DbName,
			User:         dbConfigValues.User,
			Password:     dbConfigValues.Password,
			ExtraOptions: dbConfigValues.ExtraOptions,
		}
		db := repositories.GetRepository(
			repositories.POSTGRES, dbConfig, schedulerScope.NewSubScope("database"))

		// TODO : Use admin client instead of using the execution manager
		executionManager := getExecutionManager(configuration, db, schedulerScope, applicationConfiguration)
		schedulerWorkflowExecutor := scheduler.NewWorkflowExecutor(db, executionManager, configuration, schedulerScope)
		logger.Info(context.Background(), "Successfully initialized a new scheduled workflow executor")

		schedulerWorkflowExecutor.Run()

		logger.Infof(ctx, "Flyte scheduler started successfully")
	},
}

// This is for short term and should be replaced with flyteadmin client calls.
func getExecutionManager(configuration runtimeInterfaces.Configuration, db repositories.RepositoryInterface,
	schedulerScope promutils.Scope, applicationConfiguration *runtimeInterfaces.ApplicationConfig) interfaces.ExecutionInterface {

	storeConfig := storage.GetConfig()
	serverConfig := config.GetConfig()

	execCluster := executionCluster.GetExecutionCluster(
		schedulerScope.NewSubScope("executor").NewSubScope("cluster"),
		serverConfig.KubeConfig,
		serverConfig.Master,
		configuration,
		db)

	workflowExecutor := workflowengine.NewFlytePropeller(
		applicationConfiguration.RoleNameKey,
		execCluster,
		schedulerScope.NewSubScope("executor").NewSubScope("flytepropeller"),
		configuration.NamespaceMappingConfiguration(), applicationConfiguration.EventVersion)
	logger.Info(context.Background(), "Successfully created a workflow executor engine")
	dataStorageClient, err := storage.NewDataStore(storeConfig, schedulerScope.NewSubScope("storage"))
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize storage config")
		panic(err)
	}

	publisher := notifications.NewNotificationsPublisher(*configuration.ApplicationConfiguration().GetNotificationsConfig(), schedulerScope)
	processor := notifications.NewNotificationsProcessor(*configuration.ApplicationConfiguration().GetNotificationsConfig(), schedulerScope)
	eventPublisher := notifications.NewEventsPublisher(*configuration.ApplicationConfiguration().GetExternalEventsConfig(), schedulerScope)
	go func() {
		logger.Info(context.Background(), "Started processing notifications.")
		processor.StartProcessing()
	}()


	// Configure scheduler specific remote data handler (separate from storage)
	remoteDataConfig := configuration.ApplicationConfiguration().GetRemoteDataConfig()
	urlData := data.GetRemoteDataHandler(data.RemoteDataHandlerConfig{
		CloudProvider:            remoteDataConfig.Scheme,
		SignedURLDurationMinutes: remoteDataConfig.SignedURL.DurationMinutes,
		SigningPrincipal:         remoteDataConfig.SignedURL.SigningPrincipal,
		Region:                   remoteDataConfig.Region,
		Retries:                  defaultRetries,
		RemoteDataStoreClient:    dataStorageClient,
	}).GetRemoteURLInterface()

	workflowManager := manager.NewWorkflowManager(
		db, configuration, workflowengine.NewCompiler(), dataStorageClient, applicationConfiguration.MetadataStoragePrefix,
		schedulerScope.NewSubScope("workflow_manager"))
	namedEntityManager := manager.NewNamedEntityManager(db, configuration, schedulerScope.NewSubScope("named_entity_manager"))

	executionEventWriter := eventWriter.NewWorkflowExecutionEventWriter(db, applicationConfiguration.AsyncEventsBufferSize)
	go func() {
		executionEventWriter.Run()
	}()

	executionManager := manager.NewExecutionManager(db, configuration, dataStorageClient, workflowExecutor,
		schedulerScope.NewSubScope("execution_manager"), schedulerScope.NewSubScope("scheduler_execution_metrics"),
		publisher, urlData, workflowManager, namedEntityManager, eventPublisher, executionEventWriter)

	return executionManager
}

func init() {
	RootCmd.AddCommand(schedulerRunCmd)

	// Set Keys
	labeled.SetMetricKeys(contextutils.AppNameKey, contextutils.ProjectKey, contextutils.DomainKey,
		contextutils.ExecIDKey, contextutils.WorkflowIDKey, contextutils.NodeIDKey, contextutils.TaskIDKey,
		contextutils.TaskTypeKey, common.RuntimeTypeKey, common.RuntimeVersionKey)
}
