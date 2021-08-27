package executor

import (
	"context"
	"time"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/core"
	"github.com/flyteorg/flyteadmin/scheduler/executor"
	"github.com/flyteorg/flyteadmin/scheduler/repositories"
	"github.com/flyteorg/flyteadmin/scheduler/snapshoter"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/futures"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"go.uber.org/ratelimit"
	"k8s.io/apimachinery/pkg/util/wait"
)

const snapshotWriterSleepTime = 30 * time.Second
const scheduleUpdaterSleepTime = 30 * time.Second

const snapShotVersion = 1

// ScheduledExecutor used for executing the schedules saved by the native flyte scheduler in the database.
type ScheduledExecutor struct {
	scheduler              core.Scheduler
	snapshoter             snapshoter.Persistence
	db                     repositories.SchedulerRepoInterface
	scope                  promutils.Scope
	adminServiceClient     service.AdminServiceClient
	workflowExecutorConfig *runtimeInterfaces.FlyteWorkflowExecutorConfig
}

func (w *ScheduledExecutor) Run(ctx context.Context) error {
	logger.Infof(ctx, "Flyte native scheduler started successfully")

	defer logger.Infof(ctx, "Flyte native scheduler shutdown")

	// Read snapshot from the DB.
	snapShotReader := &snapshoter.VersionedSnapshot{Version: snapShotVersion}
	snapshot, err := w.snapshoter.Read(ctx, snapShotReader)

	if err != nil {
		logger.Errorf(ctx, "unable to read the snapshot from the db due to %v. Aborting", err)
		return err
	}

	// Read all the schedules from the DB
	schedules, err := w.db.SchedulableEntityRepo().GetAll(ctx)
	if err != nil {
		logger.Errorf(ctx, "unable to read the schedules from the db due to %v. Aborting", err)
		return err
	}
	// Default to 100
	adminAPIRateLimit := 100
	if w.workflowExecutorConfig != nil {
		adminAPIRateLimit = w.workflowExecutorConfig.AdminFireReqRateLimit
	}

	// Set the rate limit on the admin
	rateLimiter := ratelimit.New(adminAPIRateLimit)

	// Set the executor to send executions to admin
	executor := executor.New(w.scope, w.adminServiceClient)

	// Create the scheduler using GoCronScheduler implementation
	gcronScheduler := core.NewGoCronScheduler(w.scope, snapshot, rateLimiter, executor)
	w.scheduler = gcronScheduler

	// Bootstrap the schedules from the snapshot
	bootStrapCtx, bootStrapCancel := context.WithCancel(ctx)
	defer bootStrapCancel()
	gcronScheduler.BootStrapSchedulesFromSnapShot(bootStrapCtx, schedules, snapshot)

	// Start the go routine to write the update schedules periodically
	updaterCtx, updaterCancel := context.WithCancel(ctx)
	defer updaterCancel()
	gcronUpdater := core.NewUpdater(w.db, gcronScheduler)
	go wait.UntilWithContext(updaterCtx, gcronUpdater.UpdateGoCronSchedules, scheduleUpdaterSleepTime)

	// Catch up simulataneously on all the schedules in the scheduler
	currTime := time.Now()
	af := futures.NewAsyncFuture(ctx, func(ctx context.Context) (interface{}, error) {
		return gcronScheduler.CatchupAll(ctx, currTime), nil
	})
	isCatchupSuccess, err := af.Get(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to get future value for catchup due to %v", err)
		return err
	}

	if isCatchupSuccess.(bool) {
		snapshotRunner := core.NewSnapshotRunner(w.snapshoter, w.scheduler)
		// Start the go routine to write the snapshot periodically
		snapshoterCtx, snapshoterCancel := context.WithCancel(ctx)
		defer snapshoterCancel()
		wait.UntilWithContext(snapshoterCtx, snapshotRunner.Run, snapshotWriterSleepTime)
		<-ctx.Done()
	}
	return nil
}

func NewScheduledExecutor(db repositories.SchedulerRepoInterface, config runtimeInterfaces.Configuration,
	scope promutils.Scope, adminServiceClient service.AdminServiceClient) ScheduledExecutor {
	return ScheduledExecutor{
		db:                     db,
		scope:                  scope,
		adminServiceClient:     adminServiceClient,
		workflowExecutorConfig: config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.GetFlyteWorkflowExecutorConfig(),
		snapshoter:             snapshoter.New(scope, db),
	}
}
