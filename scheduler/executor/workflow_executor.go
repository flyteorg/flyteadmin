package executor

import (
	"context"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	schedInterfaces "github.com/flyteorg/flyteadmin/scheduler/executor/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/repositories"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/promutils"
	"go.uber.org/ratelimit"
)

const registrationSleepTime = 30

const snapshotWriterSleepTime = 30
const checkPointerRoutineLabel = "checkpointer"
const scheduleRegisterRoutineLabel = "scheduleregister"
const catchupRoutineLabel = "catchup"

// workflowExecutor used for executing the schedules saved by the native flyte scheduler in the database.
type workflowExecutor struct {
	catchuper schedInterfaces.ScheduleCatchuper
	checkpointer schedInterfaces.ScheduleCheckPointer
	registerer schedInterfaces.ScheduleRegisterer
}

func (w *workflowExecutor) Run(ctx context.Context) error {
	// Start the go routine to catchup all the schedules
	catchupCtx, cancelCatchUp := context.WithCancel(ctx)
	defer cancelCatchUp()
	catchUpErrChan := w.catchuper.RunCatchuper(catchupCtx)

	// Start the go routine to write the snapshot periodically
	checkPointerCtx, checkPointerCancel := context.WithCancel(ctx)
	defer checkPointerCancel()
	w.checkpointer.RunCheckPointer(checkPointerCtx)

	// Start the go routine to run the schedule registerer
	scheduleRegisterCtx, scheduleRegisterCtxCancel := context.WithCancel(ctx)
	defer scheduleRegisterCtxCancel()
	w.registerer.RunRegisterer(scheduleRegisterCtx)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-catchUpErrChan:
		return err
	}
}

func NewWorkflowExecutor(db repositories.SchedulerRepoInterface, config runtimeInterfaces.Configuration,
	scope promutils.Scope, adminServiceClient service.AdminServiceClient) schedInterfaces.WorkflowExecutor {

	ctx := context.Background()
	snapShotReaderWriter := &VersionedSnapshot{version: snapShotVersion}

	// Rate limiter on admin
	workflowExecConfig := config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.GetFlyteWorkflowExecutorConfig()
	rateLimiter := ratelimit.New(workflowExecConfig.AdminFireReqRateLimit)

	goCronInterface := NewGoCron(scope)
	executionFirer := NewScheduleExecutionFirer(scope, adminServiceClient)
	checkpointer := NewScheduleCheckPointer(scope, snapShotReaderWriter, db)
	snapshot := checkpointer.ReadCheckPoint(ctx)
	catchuper := NewScheduleCatchuper(scope, snapshot, goCronInterface, rateLimiter, db, executionFirer)
	registerer := NewScheduleRegisterer(scope, goCronInterface, snapshot, snapShotReaderWriter, db, rateLimiter, executionFirer)

	return &workflowExecutor{
		catchuper: catchuper,
		checkpointer: checkpointer,
		registerer: registerer,
	}
}
