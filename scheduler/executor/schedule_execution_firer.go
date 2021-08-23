package executor

import (
	"context"
	"github.com/flyteorg/flyteadmin/scheduler/executor/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"strings"
	"time"
)

type ScheduleExecutionFirer struct {
	adminServiceClient service.AdminServiceClient
	metrics            fireMetrics
}

type fireMetrics struct {
	Scope                     promutils.Scope
	FailedExecutionCounter    prometheus.Counter
	SuccessfulExecutionCounter prometheus.Counter
}

func (w *ScheduleExecutionFirer) Fire(ctx context.Context, scheduledTime time.Time, s models.SchedulableEntity) error {

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
	executionIdentifier, err := GetExecutionIdentifier(ctx, core.Identifier{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    s.Name,
		Version: s.Version,
	}, scheduledTime)

	if err != nil {
		logger.Error(ctx, "failed to generate execution identifier for schedule %+v due to %v", s, err)
		return err
	}

	executionRequest := &admin.ExecutionCreateRequest{
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

	// Do maximum of 30 retries on failures with constant backoff factor
	opts := wait.Backoff{Factor: 1.0, Steps: 30}
	err = retry.OnError(opts,
		func(err error) bool {
			if err == nil {
				return false
			}
			// For idempotent behavior ignore the AlreadyExists error which happens if we try to schedule a launchplan
			// for execution at the same time which is already available in admin.
			// This is possible since idempotency gurantees are using the schedule time and the identifier
			if grpcError := status.Code(err); grpcError == codes.AlreadyExists {
				logger.Debugf(ctx, "duplicate schedule %+v already exists for schedule", s)
				return false
			}
			w.metrics.FailedExecutionCounter.Inc()
			logger.Error(ctx, "failed to create execution create request %+v due to %v", executionRequest, err)
			// TODO: Handle the case when admin launch plan state is archived but the schedule is active.
			// After this bug is fixed in admin https://github.com/flyteorg/flyte/issues/1354
			return true
		},
		func() error {
			_, execErr := w.adminServiceClient.CreateExecution(context.Background(), executionRequest)
			return execErr
		},
	)
	w.metrics.SuccessfulExecutionCounter.Inc()
	logger.Infof(ctx, "successfully fired the request for schedule %+v for time %v", s, scheduledTime)
	return nil
}

func NewScheduleExecutionFirer(scope promutils.Scope,
	adminServiceClient service.AdminServiceClient) interfaces.ScheduleExecutionFirer {

	return &ScheduleExecutionFirer{
		adminServiceClient: adminServiceClient,
		metrics: getFireMetrics(scope),
	}
}

func getFireMetrics(scope promutils.Scope) fireMetrics {
	return fireMetrics{
		Scope: scope,
		FailedExecutionCounter: scope.MustNewCounter("failed_execution_counter",
			"count of unsuccessful attempts to fire execution for a schedules"),
		SuccessfulExecutionCounter: scope.MustNewCounter("successful_execution_counter",
			"count of successful attempts to fire execution for a schedules"),
	}
}