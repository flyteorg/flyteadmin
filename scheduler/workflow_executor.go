package executor

import (
	"context"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/core"
	"github.com/flyteorg/flyteadmin/scheduler/executor"
	"github.com/flyteorg/flyteadmin/scheduler/repositories"
	sImpl "github.com/flyteorg/flyteadmin/scheduler/snapshoter"
	"github.com/flyteorg/flytestdlib/futures"
	"github.com/flyteorg/flytestdlib/logger"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/promutils"
	"go.uber.org/ratelimit"
	"time"
)

const snapshotWriterSleepTime = 30
const scheduleUpdaterSleepTime = 30


const snapShotVersion = 1

// workflowExecutor used for executing the schedules saved by the native flyte scheduler in the database.
type workflowExecutor struct {
	scheduler              core.Scheduler
	snapshoter             sImpl.Persistence
	db                     repositories.SchedulerRepoInterface
	scope                  promutils.Scope
	adminServiceClient     service.AdminServiceClient
	workflowExecutorConfig *runtimeInterfaces.FlyteWorkflowExecutorConfig
}

func (w *workflowExecutor) Run(ctx context.Context) error {
	// Read snapshot from the DB.
	snapShotReader := &sImpl.VersionedSnapshot{Version: snapShotVersion}
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
	adminApiRateLimit := 100
	if w.workflowExecutorConfig != nil {
		adminApiRateLimit = w.workflowExecutorConfig.AdminFireReqRateLimit
	}

	// Set the rate limit on the admin
	rateLimiter := ratelimit.New(adminApiRateLimit)

	// Set the executor to send executions to admin
	executor := executor.New(w.scope, w.adminServiceClient)

	// Create the scheduler using GoCronScheduler implementation
	gcronScheduler := core.NewGoCronScheduler(w.scope, snapshot, rateLimiter, executor)

	// Bootstrap the schedules from the snapshot
	bootStrapCtx, bootStrapCancel := context.WithCancel(ctx)
	defer bootStrapCancel()
	gcronScheduler.BootStrapSchedulesFromSnapShot(bootStrapCtx, schedules, snapshot)

	// Start the go routine to write the update schedules periodically
	updaterCtx, updaterCancel := context.WithCancel(ctx)
	defer updaterCancel()
	gcronUpdater := core.NewUpdater(w.db, gcronScheduler)
	wait.UntilWithContext(updaterCtx, gcronUpdater.UpdateGoCronSchedules, scheduleUpdaterSleepTime*time.Second)

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
		wait.UntilWithContext(snapshoterCtx, snapshotRunner.Run, snapshotWriterSleepTime*time.Second)
	} else {
		// cancel gcronUpdater and all the spawned jobs
		updaterCancel()
		bootStrapCancel()
	}
	return nil
}

func NewWorkflowExecutor(db repositories.SchedulerRepoInterface, config runtimeInterfaces.Configuration,
	scope promutils.Scope, adminServiceClient service.AdminServiceClient) workflowExecutor {

	return workflowExecutor{
		db : db,
		scope: scope,
		adminServiceClient: adminServiceClient,
		workflowExecutorConfig: config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.GetFlyteWorkflowExecutorConfig(),
	}
}
