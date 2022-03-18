package adminservice

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/auth"
	"runtime/debug"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"

	eventWriter "github.com/flyteorg/flyteadmin/pkg/async/events/implementations"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule"
	"github.com/flyteorg/flyteadmin/pkg/data"
	executionCluster "github.com/flyteorg/flyteadmin/pkg/executioncluster/impl"
	manager "github.com/flyteorg/flyteadmin/pkg/manager/impl"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine"
	workflowengineImpl "github.com/flyteorg/flyteadmin/pkg/workflowengine/impl"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/profutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
)

type AdminService struct {
	service.UnimplementedAdminServiceServer
	TaskManager          interfaces.TaskInterface
	WorkflowManager      interfaces.WorkflowInterface
	LaunchPlanManager    interfaces.LaunchPlanInterface
	ExecutionManager     interfaces.ExecutionInterface
	NodeExecutionManager interfaces.NodeExecutionInterface
	TaskExecutionManager interfaces.TaskExecutionInterface
	ProjectManager       interfaces.ProjectInterface
	ResourceManager      interfaces.ResourceInterface
	NamedEntityManager   interfaces.NamedEntityInterface
	VersionManager       interfaces.VersionInterface
	Metrics              AdminMetrics
}

// Intercepts all admin requests to handle panics during execution.
func (m *AdminService) interceptPanic(ctx context.Context, request proto.Message) {
	err := recover()
	if err == nil {
		return
	}

	m.Metrics.PanicCounter.Inc()
	logger.Fatalf(ctx, "panic-ed for request: [%+v] with err: %v with Stack: %v", request, err, string(debug.Stack()))
}

const defaultRetries = 3

func NewAdminServer(ctx context.Context, kubeConfig, master string) *AdminService {
	configuration := runtime.NewConfigurationProvider()
	applicationConfiguration := configuration.ApplicationConfiguration().GetTopLevelConfig()

	adminScope := promutils.NewScope(applicationConfiguration.GetMetricsScope()).NewSubScope("admin")
	panicCounter := adminScope.MustNewCounter("initialization_panic",
		"panics encountered initializing the admin service")

	defer func() {
		if err := recover(); err != nil {
			panicCounter.Inc()
			logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
	logConfig := logger.GetConfig()

	db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	dbScope := adminScope.NewSubScope("database")
	repo := repositories.NewGormRepo(
		db, errors.NewPostgresErrorTransformer(adminScope.NewSubScope("errors")), dbScope)
	storeConfig := storage.GetConfig()
	execCluster := executionCluster.GetExecutionCluster(
		adminScope.NewSubScope("executor").NewSubScope("cluster"),
		kubeConfig,
		master,
		configuration,
		repo)
	workflowBuilder := workflowengineImpl.NewFlyteWorkflowBuilder(
		adminScope.NewSubScope("builder").NewSubScope("flytepropeller"))
	workflowExecutor := workflowengineImpl.NewK8sWorkflowExecutor(execCluster, workflowBuilder)
	logger.Info(ctx, "Successfully created a workflow executor engine")
	workflowengine.GetRegistry().RegisterDefault(workflowExecutor)

	logger.Warnf(ctx, "**Registering blanket auth server")
	auth.GetInterceptorProvider().RegisterDefault(auth.BlanketAuthorization)

	dataStorageClient, err := storage.NewDataStore(storeConfig, adminScope.NewSubScope("storage"))
	if err != nil {
		logger.Error(ctx, "Failed to initialize storage config")
		panic(err)
	}

	publisher := notifications.NewNotificationsPublisher(*configuration.ApplicationConfiguration().GetNotificationsConfig(), adminScope)
	processor := notifications.NewNotificationsProcessor(*configuration.ApplicationConfiguration().GetNotificationsConfig(), adminScope)
	eventPublisher := notifications.NewEventsPublisher(*configuration.ApplicationConfiguration().GetExternalEventsConfig(), adminScope)
	go func() {
		logger.Info(ctx, "Started processing notifications.")
		processor.StartProcessing()
	}()

	// Configure workflow scheduler async processes.
	schedulerConfig := configuration.ApplicationConfiguration().GetSchedulerConfig()
	workflowScheduler := schedule.NewWorkflowScheduler(repo, schedule.WorkflowSchedulerConfig{
		Retries:         defaultRetries,
		SchedulerConfig: *schedulerConfig,
		Scope:           adminScope,
	})

	eventScheduler := workflowScheduler.GetEventScheduler()
	launchPlanManager := manager.NewLaunchPlanManager(
		repo, configuration, eventScheduler, adminScope.NewSubScope("launch_plan_manager"))

	// Configure admin-specific remote data handler (separate from storage)
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
		repo, configuration, workflowengineImpl.NewCompiler(), dataStorageClient, applicationConfiguration.GetMetadataStoragePrefix(),
		adminScope.NewSubScope("workflow_manager"))
	namedEntityManager := manager.NewNamedEntityManager(repo, configuration, adminScope.NewSubScope("named_entity_manager"))

	executionEventWriter := eventWriter.NewWorkflowExecutionEventWriter(repo, applicationConfiguration.GetAsyncEventsBufferSize())
	go func() {
		executionEventWriter.Run()
	}()

	executionManager := manager.NewExecutionManager(repo, configuration, dataStorageClient,
		adminScope.NewSubScope("execution_manager"), adminScope.NewSubScope("user_execution_metrics"),
		publisher, urlData, workflowManager, namedEntityManager, eventPublisher, executionEventWriter)
	versionManager := manager.NewVersionManager()

	scheduledWorkflowExecutor := workflowScheduler.GetWorkflowExecutor(executionManager, launchPlanManager)
	logger.Info(ctx, "Successfully initialized a new scheduled workflow executor")
	go func() {
		logger.Info(ctx, "Starting the scheduled workflow executor")
		scheduledWorkflowExecutor.Run()
	}()

	// Serve profiling endpoints.
	go func() {
		err := profutils.StartProfilingServerWithDefaultHandlers(
			ctx, applicationConfiguration.GetProfilerPort(), nil)
		if err != nil {
			logger.Panicf(ctx, "Failed to Start profiling and Metrics server. Error, %v", err)
		}
	}()

	nodeExecutionEventWriter := eventWriter.NewNodeExecutionEventWriter(repo, applicationConfiguration.GetAsyncEventsBufferSize())
	go func() {
		nodeExecutionEventWriter.Run()
	}()

	logger.Info(ctx, "Initializing a new AdminService")
	return &AdminService{
		TaskManager: manager.NewTaskManager(repo, configuration, workflowengineImpl.NewCompiler(),
			adminScope.NewSubScope("task_manager")),
		WorkflowManager:    workflowManager,
		LaunchPlanManager:  launchPlanManager,
		ExecutionManager:   executionManager,
		NamedEntityManager: namedEntityManager,
		VersionManager:     versionManager,
		NodeExecutionManager: manager.NewNodeExecutionManager(repo, configuration, applicationConfiguration.GetMetadataStoragePrefix(), dataStorageClient,
			adminScope.NewSubScope("node_execution_manager"), urlData, eventPublisher, nodeExecutionEventWriter),
		TaskExecutionManager: manager.NewTaskExecutionManager(repo, configuration, dataStorageClient,
			adminScope.NewSubScope("task_execution_manager"), urlData, eventPublisher),
		ProjectManager:  manager.NewProjectManager(repo, configuration),
		ResourceManager: resources.NewResourceManager(repo, configuration.ApplicationConfiguration()),
		Metrics:         InitMetrics(adminScope),
	}
}
