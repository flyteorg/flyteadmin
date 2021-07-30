package flytescheduler

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	mgInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/gogf/gf/os/gcron"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"sync"

	"github.com/robfig/cron"

	"time"
)

const executorSleepTime = 30
const backOffSleepTime = 60

type workflowExecutor struct {
	db               repositories.RepositoryInterface
	config           runtimeInterfaces.Configuration
	executionManager mgInterfaces.ExecutionInterface
	initialized      bool
}

type SyncerModuleCommandWithPayload struct {
	SyncerModuleCommand
	scheduleName string
}

type SyncerModuleCommand int

const (
	Incr SyncerModuleCommand = iota
	Rm
	Decr
)


var mapOfSchedules sync.Map

type ScheduleRunTimeData struct {
	ranInLastTick bool
	lastTickScheduleTs time.Time
}

var SyncerModuleChannel chan SyncerModuleCommandWithPayload

// Syncer module variables

func CheckPointState(repo repositories.RepositoryInterface) {
	mapOfSchedulesLocal := map[string]bool{}
	oneRoundSchedulesMap := map[string]bool{}
	for command := range SyncerModuleChannel {
		switch command.SyncerModuleCommand {
		case Incr:
			mapOfSchedulesLocal[command.scheduleName] = true
			oneRoundSchedulesMap[command.scheduleName] = true
		case Rm:
			delete(mapOfSchedulesLocal, command.scheduleName)
			delete(oneRoundSchedulesMap, command.scheduleName)
		case Decr:
			delete(oneRoundSchedulesMap, command.scheduleName)
			if len(oneRoundSchedulesMap) == 0 {
				// Flush all the records to the DB with a checkpoint
				checkPoint(time.Now(), repo)
				// Copy mapOfSchedulesLocal for next round
				for key, value := range mapOfSchedulesLocal {
					oneRoundSchedulesMap[key] = value
				}
			}
		}
	}
}

func checkPoint(checkPointTime time.Time, repo repositories.RepositoryInterface) {
	ctx := context.Background()
	err := repo.ScheduleCheckPointRepo().Update(ctx, models.ScheduleCheckPoint{CheckPointTime: &checkPointTime})
	if err != nil {
		logger.Errorf(ctx, "Failed to update checkpoint time  %v due to %v", checkPointTime, err)
	}
}

func (w *workflowExecutor) Run() {
	ctx := context.Background()
	SyncerModuleChannel = make(chan SyncerModuleCommandWithPayload)

	c, err := w.db.SchedulableEntityRepo().GetAllActive(ctx)
	if err != nil {
		panic(fmt.Errorf("unable to run the workflow executor after reading the schedules due to %v", err))
	}

	go CheckPointState(w.db)
	defer logger.Infof(ctx, "Exiting Workflow executor")
	for true {
		for _, s := range c.Entities {
			// e.CronExpression
			// Run the 5 secs pattern
			funcRef := func() {
				nameOfSchedule := getNameForSchedule(s)
				// This is the case where the schedule was deactivated from admin
				// Worst case if this check succeeds and then schedule gets deactivated then it would run atmost once additionally after deactivation.
				r,ok := mapOfSchedules.Load(nameOfSchedule)
				if !ok {
					logger.Errorf(ctx,"scheduler job %v doesn't exist as it must have been deleted", nameOfSchedule)
					return
				}
				scheduleRunTimeData,ok := r.(ScheduleRunTimeData)
				if !ok {
					logger.Errorf(ctx,"incorrect type of %T stored in runtime data for the schedule", scheduleRunTimeData)
					return
				}
				timeMarker := getTimeMarkerForLastWorkflow(scheduleRunTimeData, w.config)
				scheduleTime, err := getScheduledTime(s.CronExpression, timeMarker) // Next schedule time from the timeMarker
				if err != nil {
					logger.Errorf(ctx, "failed to get the next schedule time using cron expression %v with marker %v due to %v", s.CronExpression, timeMarker, err)
					// Send a decrement command on the channel
					SyncerModuleChannel <- SyncerModuleCommandWithPayload{Decr, nameOfSchedule}
					return
				}
				literalsInputMap := map[string]*core.Literal{}
				literalsInputMap[s.KickoffTimeInputArg] = &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Datetime{
										Datetime: timestamppb.New(scheduleTime),
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
							ScheduledAt: timestamppb.New(scheduleTime),
						},
						// No dynamic notifications are configured either.
					},
					// No additional inputs beyond the to-be-filled-out kickoff time arg are specified.
					Inputs: &core.LiteralMap{
						Literals: literalsInputMap,
					},
				}

				_, err = w.executionManager.CreateExecution(context.Background(), executionRequest, scheduleTime)
				if err != nil {
					logger.Error(ctx, "failed to create  execution create  request due to %v", err)
				}
				// Same check here before updating the timestamp in the mapOfSchedules.
				// This checks if the nameOfSchedule exists and if not just return without updating the run timestamp
				if _,ok := mapOfSchedules.Load(nameOfSchedule); !ok {
					logger.Errorf(ctx,"scheduler job %v doesn't exist as it must have been deleted", nameOfSchedule)
					return
				}
				// Send a decrement command on the channel
				SyncerModuleChannel <- SyncerModuleCommandWithPayload{Decr, nameOfSchedule}
				mapOfSchedules.Store(nameOfSchedule, ScheduleRunTimeData{true, scheduleTime})
			}
			nameOfSchedule := getNameForSchedule(s)
			if s.Active {
				_, err = gcron.AddSingleton(s.CronExpression, funcRef, nameOfSchedule)
				if err != nil {
					if strings.HasSuffix(err.Error(), "already exists") {
						// Do nothing here and ignore the error
						return
					}
					logger.Errorf(ctx, "failed to add cron schedule %v due to %v", s, err)
				} else {
					SyncerModuleChannel <- SyncerModuleCommandWithPayload{Incr, nameOfSchedule}
					mapOfSchedules.Store(nameOfSchedule, ScheduleRunTimeData{})
				}
			} else {
				// Send a decrement command on the channel since we deleted a schedule
				// Only send a decrement if it didn't run in last tick to avoid double decrement error.
				val, ok := mapOfSchedules.Load(nameOfSchedule)
				if ok {
					if runTimeData,o := val.(ScheduleRunTimeData); o {
						if !runTimeData.ranInLastTick {
							SyncerModuleChannel <- SyncerModuleCommandWithPayload{Rm, nameOfSchedule}
						}
					}
					mapOfSchedules.Delete(nameOfSchedule)
				}
				gcron.Remove(nameOfSchedule)
			}
		}
		time.Sleep(executorSleepTime * time.Second)
		c, err = w.db.SchedulableEntityRepo().GetAllActive(ctx)
		if err != nil {
			logger.Errorf(ctx, "going to sleep additional %v backoff time due to DB error %v", backOffSleepTime, err)
			time.Sleep(backOffSleepTime * time.Second)
		}
	}
}

func getNameForSchedule(schedule models.SchedulableEntity) string {
	return fmt.Sprintf("Project: %v Domain: %v Name: %v Version: %v", schedule.Project, schedule.Domain, schedule.Name, schedule.Version)
}

func getTimeMarkerForLastWorkflow(s ScheduleRunTimeData, config runtimeInterfaces.Configuration) time.Time {
	if s.ranInLastTick {
		return s.lastTickScheduleTs
	}
	if config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.FlyteWorkflowExecutorConfig.SchedulerEpochTime != nil {
		return *config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.FlyteWorkflowExecutorConfig.SchedulerEpochTime
	}
	return time.Now()
}

func getScheduledTime(cronString string, fromTime time.Time) (time.Time, error) {
	var secondParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	sched, err := secondParser.Parse(cronString)

	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(fromTime), nil
}

func (w *workflowExecutor) Stop() error {
	return nil
}

func NewWorkflowExecutor(db repositories.RepositoryInterface, executionManager mgInterfaces.ExecutionInterface,
	config runtimeInterfaces.Configuration) interfaces.WorkflowExecutor {
	ctx := context.Background()
	checkPointData, err := db.ScheduleCheckPointRepo().Get(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to read checkpoint data due to %v", err)
	} else {
		if checkPointData.CheckPointTime != nil {
			config.ApplicationConfiguration().GetSchedulerConfig().WorkflowExecutorConfig.FlyteWorkflowExecutorConfig.SchedulerEpochTime = checkPointData.CheckPointTime
		}
	}
	return &workflowExecutor{db: db, executionManager: executionManager, config: config}
}
