package schedule

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/flytescheduler"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/scheduler"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async"

	gizmoConfig "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsSchedule "github.com/flyteorg/flyteadmin/pkg/async/schedule/aws"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/noop"
	"github.com/flyteorg/flyteadmin/pkg/common"
	managerInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

type WorkflowSchedulerConfig struct {
	Retries         int
	SchedulerConfig runtimeInterfaces.SchedulerConfig
	Scope           promutils.Scope
}

type WorkflowScheduler interface {
	GetEventScheduler() interfaces.EventScheduler
	GetWorkflowExecutor(db repositories.SchedulerRepoInterface, executionManager managerInterfaces.ExecutionInterface,
		launchPlanManager managerInterfaces.LaunchPlanInterface, config runtimeInterfaces.Configuration) interfaces.WorkflowExecutor
}

type workflowScheduler struct {
	cfg              WorkflowSchedulerConfig
	eventScheduler   interfaces.EventScheduler
	workflowExecutor interfaces.WorkflowExecutor
}

func (w *workflowScheduler) GetEventScheduler() interfaces.EventScheduler {
	return w.eventScheduler
}

func (w *workflowScheduler) GetWorkflowExecutor(
	db repositories.SchedulerRepoInterface,
	executionManager managerInterfaces.ExecutionInterface,
	launchPlanManager managerInterfaces.LaunchPlanInterface, config runtimeInterfaces.Configuration) interfaces.WorkflowExecutor {
	if w.workflowExecutor == nil {
		switch w.cfg.SchedulerConfig.WorkflowExecutorConfig.Scheme {
		case common.AWS:
			sqsConfig := gizmoConfig.SQSConfig{
				QueueName:           w.cfg.SchedulerConfig.WorkflowExecutorConfig.ScheduleQueueName,
				QueueOwnerAccountID: w.cfg.SchedulerConfig.WorkflowExecutorConfig.AccountID,
			}
			sqsConfig.Region = w.cfg.SchedulerConfig.WorkflowExecutorConfig.Region
			w.workflowExecutor = awsSchedule.NewWorkflowExecutor(
				sqsConfig, w.cfg.SchedulerConfig, executionManager, launchPlanManager, w.cfg.Scope.NewSubScope("workflow_executor"))
			break
		case common.Local:
			logger.Infof(context.Background(),
				"Using default flyte workflow executor implementation")
			w.workflowExecutor = scheduler.NewWorkflowExecutor(db, executionManager, config)
		default:
			logger.Infof(context.Background(),
				"Using default noop workflow executor implementation for cloud provider type [%s]",
				w.cfg.SchedulerConfig.EventSchedulerConfig.Scheme)
			w.workflowExecutor = noop.NewWorkflowExecutor()
		}
	}
	return w.workflowExecutor
}

func NewWorkflowScheduler(db repositories.RepositoryInterface, cfg WorkflowSchedulerConfig) WorkflowScheduler {
	var eventScheduler interfaces.EventScheduler
	var workflowExecutor interfaces.WorkflowExecutor

	switch cfg.SchedulerConfig.EventSchedulerConfig.Scheme {
	case common.AWS:
		awsConfig := aws.NewConfig().WithRegion(cfg.SchedulerConfig.WorkflowExecutorConfig.Region).WithMaxRetries(cfg.Retries)
		var sess *session.Session
		var err error
		err = async.Retry(cfg.SchedulerConfig.ReconnectAttempts,
			time.Duration(cfg.SchedulerConfig.ReconnectDelaySeconds)*time.Second, func() error {
				sess, err = session.NewSession(awsConfig)
				if err != nil {
					logger.Warnf(context.TODO(), "Failed to initialize new event scheduler with aws config: [%+v] and err: %v", awsConfig, err)
				}
				return err
			})

		if err != nil {
			panic(err)
		}
		eventScheduler = awsSchedule.NewCloudWatchScheduler(
			cfg.SchedulerConfig.EventSchedulerConfig.ScheduleRole, cfg.SchedulerConfig.EventSchedulerConfig.TargetName, sess, awsConfig,
			cfg.Scope.NewSubScope("cloudwatch_scheduler"))
	case common.Local:
		logger.Infof(context.Background(),
			"Using default flyte scheduler implementation")
		eventScheduler = flytescheduler.NewFlyteScheduler(db)
	default:
		logger.Infof(context.Background(),
			"Using default noop event scheduler implementation for cloud provider type [%s]",
			cfg.SchedulerConfig.EventSchedulerConfig.Scheme)
		eventScheduler = noop.NewNoopEventScheduler()
	}
	return &workflowScheduler{
		cfg:              cfg,
		eventScheduler:   eventScheduler,
		workflowExecutor: workflowExecutor,
	}
}
