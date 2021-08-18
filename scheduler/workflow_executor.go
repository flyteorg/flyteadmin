package scheduler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	mgInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	schedInterfaces "github.com/flyteorg/flyteadmin/scheduler/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/robfig/cron"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

const executorSleepTime = 30

const snapshotWriterSleepTime = 30
const backOffSleepTime = 60

// workflowExecutor used for executing the schedules saved by the native flyte scheduler in the database.
type workflowExecutor struct {
	db                   repositories.SchedulerRepoInterface
	config               runtimeInterfaces.Configuration
	executionManager     mgInterfaces.ExecutionInterface
	snapshot             schedInterfaces.Snapshoter
	snapShotReaderWriter schedInterfaces.SnapshotReaderWriter
	goGfInterface        schedInterfaces.GoCronWrapper
	rateLimiter          ratelimit.Limiter
}

func (w *workflowExecutor) CheckPointState(ctx context.Context) {
	for true {
		var bytesArray []byte
		f := bytes.NewBuffer(bytesArray)
		// Only write if the snapshot has contents and not equal to the previous snapshot
		if !w.snapshot.IsEmpty() {
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
		time.Sleep(snapshotWriterSleepTime * time.Second)
	}
}

func (w *workflowExecutor) CatchUpAllSchedules(ctx context.Context, schedules []models.SchedulableEntity, toTime time.Time) error {
	logger.Debugf(ctx, "catching up [%v] schedules until time %v", len(schedules), toTime)
	for _, s := range schedules {
		fromTime := time.Now()
		// If the schedule is not active, don't do anything else use the updateAt timestamp to find when the schedule became active
		// We support catchup only from the last active state
		// i.e if the schedule was Active(t1)-Archive(t2)-Active(t3)-Archive(t4)-Active-(t5)
		// And if the scheduler was down during t1-t5 , then when it comes back up it would use t5 timestamp
		// to catch up until the current timestamp
		// Here the assumption is updateAt timestamp changes for active/inactive transitions and no other changes.
		if !*s.Active {
			logger.Debugf(ctx, "schedule %+v was inactive during catchup", s)
			continue
		} else {
			fromTime = s.UpdatedAt
		}
		nameOfSchedule := GetScheduleName(s)
		lastT := w.snapshot.GetLastExecutionTime(nameOfSchedule)
		if !lastT.IsZero() && lastT.After(s.UpdatedAt) {
			fromTime = lastT
		}
		logger.Debugf(ctx, "catching up schedule %+v from %v to %v", s, fromTime, toTime)
		err := w.CatchUpSingleSchedule(ctx, s, fromTime, toTime)
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "caught up successfully on the schedule %+v from %v to %v", s, fromTime, toTime)
	}
	return nil
}

func (w *workflowExecutor) CatchUpSingleSchedule(ctx context.Context, s models.SchedulableEntity, fromTime time.Time, toTime time.Time) error {
	nameOfSchedule := GetScheduleName(s)
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
			logger.Errorf(ctx, "unable to fire the schedule %+v at %v time due to %v", s, catchupTime, err)
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

	// Making the identifier deterministic using the hash of the identifier and scheduled time
	executionIdentifier, err := GetExecutionIdentifier(core.Identifier{
		Project: s.Project,
		Domain: s.Domain,
		Name: s.Name,
		Version: s.Version,
	}, scheduledTime)

	if err != nil {
		logger.Error(ctx, "failed to generate execution identifier for schedule %+v due to %v", s, err)
		return err
	}

	executionRequest := admin.ExecutionCreateRequest{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    "f" + strings.ReplaceAll(executionIdentifier.String(), "-", "")[:19],
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
	if !*s.Active {
		// no longer active
		logger.Debugf(ctx, "schedule %+v is no longer active", s)
		return nil
	}
	_, err = executionManager.CreateExecution(context.Background(), executionRequest, scheduledTime)
	if err != nil {
		// For idempotent behavior ignore the AlreadyExists error which happens if we try to schedule a launchplan
		// for execution at the same time which is already available in admin.
		// This is possible since idempotency gurantees are using the schedule time and the identifier
		if grpcError := status.Code(err); grpcError == codes.AlreadyExists {
			logger.Debugf(ctx, "duplicate schedule %+v already exists for schedule", s)
			return nil
		}
		logger.Error(ctx, "failed to create execution create request %+v due to %v", executionRequest, err)
		return err
	}
	return nil
}

func (w *workflowExecutor) Run() {
	ctx := context.Background()
	schedules, err := w.db.SchedulableEntityRepo().GetAll(ctx)
	if err != nil {
		panic(fmt.Errorf("unable to run the workflow executor after reading the schedules due to %v", err))
	}
	catchUpTill := time.Now()
	err = w.CatchUpAllSchedules(ctx, schedules, catchUpTill)
	if err != nil {
		panic(fmt.Errorf("unable to catch up on the schedules due to %v", err))
	}

	// Start the go routine to write the snapshot map to the configured snapshot file
	go w.CheckPointState(ctx)

	defer logger.Infof(ctx, "Exiting Workflow executor")
	for true {
		for _, s := range schedules {
			funcRef := func(ctx context.Context, schedule models.SchedulableEntity, scheduleTime time.Time) {
				// If the schedule has been deactivated and then the inflight schedules can stop
				if !*schedule.Active {
					return
				}
				nameOfSchedule := GetScheduleName(schedule)
				_ = w.rateLimiter.Take()
				err := fire(ctx, w.executionManager, scheduleTime, schedule)
				if err != nil {
					logger.Errorf(ctx, "unable to fire the schedule %+v at %v time due to %v", s, scheduleTime, err)
					return
				} else {
					w.snapshot.UpdateLastExecutionTime(nameOfSchedule, scheduleTime)
				}
			}

			// Register or deregister the schedule from the scheduler
			if !*s.Active {
				w.goGfInterface.DeRegister(ctx, s)
			} else {
				err := w.goGfInterface.Register(ctx, s, funcRef)
				if err != nil {
					logger.Errorf(ctx, "unable to register the schedule %+v due to %v", s, err)
				}
			}
		}
		time.Sleep(executorSleepTime * time.Second)
		schedules, err = w.db.SchedulableEntityRepo().GetAll(ctx)
		if err != nil {
			logger.Errorf(ctx, "going to sleep additional %v backoff time due to DB error %v", backOffSleepTime, err)
			time.Sleep(backOffSleepTime * time.Second)
		}
	}
}

func (w *workflowExecutor) Stop() error {
	return nil
}

func NewWorkflowExecutor(db repositories.SchedulerRepoInterface, executionManager mgInterfaces.ExecutionInterface,
	config runtimeInterfaces.Configuration) interfaces.WorkflowExecutor {
	ctx := context.Background()
	workflowExecConfig := config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.GetFlyteWorkflowExecutorConfig()
	snapShotReaderWriter := VersionedSnapshot{version: workflowExecConfig.SnapshotVersion}
	rateLimiter := ratelimit.New(workflowExecConfig.AdminFireReqRateLimit)
	snapshot := readSnapShot(ctx, db, workflowExecConfig.SnapshotVersion)
	cron := cron.New()
	cron.Start()
	return &workflowExecutor{db: db, executionManager: executionManager, config: config, snapshot: snapshot,
		snapShotReaderWriter: &snapShotReaderWriter,
		goGfInterface:        GoCron{jobsMap: map[string]schedInterfaces.GoCronJobWrapper{}, c: cron},
		rateLimiter:          rateLimiter}
}

func readSnapShot(ctx context.Context, db repositories.SchedulerRepoInterface, version int) schedInterfaces.Snapshoter {
	var snapshot schedInterfaces.Snapshoter
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
