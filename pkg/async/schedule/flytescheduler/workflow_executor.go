package flytescheduler

import (
	"bytes"
	"context"
	"fmt"
	flyteSchedulerInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/flytescheduler/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	mgInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/google/uuid"
	"go.uber.org/ratelimit"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

const executorSleepTime = 30

const snapshotWriterSleepTime = 30
const backOffSleepTime = 60

// workflowExecutor used for executing the schedules saved by the native flyte scheduler in the database.
type workflowExecutor struct {
	db                      repositories.RepositoryInterface
	config                  runtimeInterfaces.Configuration
	executionManager        mgInterfaces.ExecutionInterface
	snapshot                flyteSchedulerInterfaces.Snapshot
	snapShotReaderWriter    flyteSchedulerInterfaces.SnapshotReaderWriter
	goGfInterface           flyteSchedulerInterfaces.GoGFWrapper
	executionRequestsPerSec int
	rateLimiter             ratelimit.Limiter
}

func (w *workflowExecutor) CheckPointState(ctx context.Context) {
	var prevSnapshot flyteSchedulerInterfaces.Snapshot
	for true {
		var bytesArray []byte
		f := bytes.NewBuffer(bytesArray)
		// Only write if the snapshot has contents and not equal to the previous snapshot
		if !w.snapshot.IsEmpty() && !w.snapshot.AreEqual(prevSnapshot) {
			err := w.snapShotReaderWriter.WriteSnapshot(f, w.snapshot)
			// Just log the error
			if err != nil {
				logger.Errorf(ctx, "unable to write the snapshot to buffer due to %v", err)
			}
			err = w.db.ScheduleEntitiesSnapshotRepo().CreateSnapShot(ctx, models.ScheduleEntitiesSnapshot{
				Snapshot: f.Bytes(),
			})
			if err != nil {
				logger.Errorf(ctx, "unable to save the snapshot to the database due to %v", err)
			}
		}
		prevSnapshot = w.snapshot
		time.Sleep(snapshotWriterSleepTime * time.Second)
	}
}

func (w *workflowExecutor) CatchUpAllSchedules(ctx context.Context, schedules []models.SchedulableEntity, toTime time.Time) error {
	for _, s := range schedules {
		err := w.CatchUpSingleSchedule(ctx, s, toTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *workflowExecutor) CatchUpSingleSchedule(ctx context.Context, s models.SchedulableEntity, toTime time.Time) error {
	nameOfSchedule := w.goGfInterface.GetScheduleName(s)
	lastExecTime := w.snapshot.GetLastExecutionTime(nameOfSchedule)
	// Get the jitter value
	workflowExecConfig := w.config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.GetFlyteWorkflowExecutorConfig()
	jitter := time.Duration(workflowExecConfig.JitterValue) * time.Second
	schedulerEpochTime := workflowExecConfig.SchedulerEpochTime

	var fromTime time.Time
	// If the scheduler epoch time ise set then use that to catch up till the current time
	// else check the last exec time for the schedule and if that also is not set then use current time with
	// configured jitter
	if !schedulerEpochTime.IsZero() {
		fromTime = schedulerEpochTime
	} else {
		if !lastExecTime.IsZero() {
			fromTime = lastExecTime
		} else {
			fromTime = time.Now().Add(-jitter)
		}
	}

	var catchUpTimes []time.Time
	var err error
	catchUpTimes, err = w.goGfInterface.GetCatchUpTimes(s, fromTime, toTime)
	if err != nil {
		return err
	}
	var catchupTime time.Time
	for _, catchupTime = range catchUpTimes {
		_ = w.rateLimiter.Take()
		err := fire(ctx, w.executionManager, catchupTime, s)
		if err != nil {
			logger.Errorf(ctx, "unable to fire the schedule %v at %v time due to %v", s, catchupTime, err)
			return err
		} else {
			w.snapshot.UpdateLastExecutionTime(nameOfSchedule, catchupTime)
		}
	}

	return nil
}

func fire(ctx context.Context, executionManager mgInterfaces.ExecutionInterface, scheduledTime time.Time,
	s models.SchedulableEntity) error {

	literalsInputMap := map[string]*core.Literal{}
	literalsInputMap[s.KickoffTimeInputArg] = &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Datetime{
							Datetime: timestamppb.New(scheduledTime),
						},
					},
				},
			},
		},
	}

	executionRequest := admin.ExecutionCreateRequest{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    "f" + strings.ReplaceAll(uuid.New().String(), "-", "")[:19],
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      s.Project,
				Domain:       s.Domain,
				Name:         s.Name,
				Version:      s.Version,
			},
			Metadata: &admin.ExecutionMetadata{
				Mode:        admin.ExecutionMetadata_SCHEDULED,
				ScheduledAt: timestamppb.New(scheduledTime),
			},
			// No dynamic notifications are configured either.
		},
		// No additional inputs beyond the to-be-filled-out kickoff time arg are specified.
		Inputs: &core.LiteralMap{
			Literals: literalsInputMap,
		},
	}

	_, err := executionManager.CreateExecution(context.Background(), executionRequest, scheduledTime)
	if err != nil {
		logger.Error(ctx, "failed to create  execution create  request due to %v", err)
		return err
	}
	return nil
}

func (w *workflowExecutor) Run() {
	ctx := context.Background()
	c, err := w.db.SchedulableEntityRepo().GetAll(ctx)
	if err != nil {
		panic(fmt.Errorf("unable to run the workflow executor after reading the schedules due to %v", err))
	}
	// Start the go routine to write the snapshot map to the configured snapshot file
	go w.CheckPointState(ctx)
	catchUpTill := time.Now()
	err = w.CatchUpAllSchedules(ctx, c.Entities, catchUpTill)
	if err != nil {
		panic(fmt.Errorf("unable to catch up on the schedules due to %v", err))
	}

	defer logger.Infof(ctx, "Exiting Workflow executor")
	for true {
		for _, s := range c.Entities {
			funcRef := func() {
				// If the schedule has been deactivated and then the inflight schedules can stop at the beggining
				if !*s.Active {
					return
				}
				nameOfSchedule := w.goGfInterface.GetScheduleName(s)
				lastT := w.snapshot.GetLastExecutionTime(nameOfSchedule)
				n := time.Now()
				err = w.CatchUpSingleSchedule(ctx, s, n)
				if err != nil {
					logger.Errorf(ctx, "unable to catch on schedule %v from lastT %v to %v with jitter configured due to %v", s.Name, lastT, n, err)
				}
			}
			err := w.goGfInterface.Register(ctx, s, funcRef)
			if err != nil {
				logger.Errorf(ctx, "unable to register the schedule %v due to %v", s, err)
			}
		}
		time.Sleep(executorSleepTime * time.Second)
		c, err = w.db.SchedulableEntityRepo().GetAll(ctx)
		if err != nil {
			logger.Errorf(ctx, "going to sleep additional %v backoff time due to DB error %v", backOffSleepTime, err)
			time.Sleep(backOffSleepTime * time.Second)
		}
	}
}

func (w *workflowExecutor) Stop() error {
	return nil
}

func NewWorkflowExecutor(db repositories.RepositoryInterface, executionManager mgInterfaces.ExecutionInterface,
	config runtimeInterfaces.Configuration) interfaces.WorkflowExecutor {
	ctx := context.Background()
	workflowExecConfig := config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.GetFlyteWorkflowExecutorConfig()
	snapShotReaderWriter := VersionedSnapshot{version: workflowExecConfig.SnapshotVersion}
	rateLimiter := ratelimit.New(workflowExecConfig.AdminFireReqRateLimit)
	snapshot := readSnapShot(ctx, db, workflowExecConfig.SnapshotVersion)
	return &workflowExecutor{db: db, executionManager: executionManager, config: config, snapshot: snapshot,
		snapShotReaderWriter: &snapShotReaderWriter,
		goGfInterface:        GoGF{},
		rateLimiter:          rateLimiter}
}

func readSnapShot(ctx context.Context, db repositories.RepositoryInterface, version int) flyteSchedulerInterfaces.Snapshot {
	var snapshot flyteSchedulerInterfaces.Snapshot
	scheduleEntitiesSnapShot, err := db.ScheduleEntitiesSnapshotRepo().GetLatestSnapShot(ctx)
	// Just log the error but dont interrupt the startup of the scheduler
	if err != nil {
		logger.Errorf(ctx, "unable to read the snapshot from the DB due to %v", err)
	} else {
		f := bytes.NewReader(scheduleEntitiesSnapShot.Snapshot)
		snapShotReaderWriter := VersionedSnapshot{version: version}
		snapshot, err = snapShotReaderWriter.ReadSnapshot(f)
		// Similarly just log the error but dont interrupt the startup of the scheduler
		if err != nil {
			logger.Errorf(ctx, "unable to construct the snapshot struct from the file due to %v", err)
			return &SnapshotV1{LastTimes: map[string]time.Time{}}
		}
		return snapshot
	}
	return &SnapshotV1{LastTimes: map[string]time.Time{}}
}
