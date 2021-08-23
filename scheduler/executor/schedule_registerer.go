package executor

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/scheduler/executor/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/repositories"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/ratelimit"
	"k8s.io/apimachinery/pkg/util/wait"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

type scheduleRegistererMetrics struct {
	Scope                            promutils.Scope
	ScheduleRegistrationPanicCounter prometheus.Counter
	ScheduleRegistrationFailure      prometheus.Counter
}

type ScheduleRegisterer struct {
	metrics              scheduleRegistererMetrics
	snapshoter           interfaces.Snapshoter
	snapShotReaderWriter interfaces.SnapshotReaderWriter
	goCronInterface      interfaces.GoCronWrapper
	rateLimiter          ratelimit.Limiter
	db                   repositories.SchedulerRepoInterface
	executionFirer       interfaces.ScheduleExecutionFirer
}

func (w *ScheduleRegisterer) RunRegisterer(ctx context.Context) {
	scheduleRegisterCtxWithLabel := contextutils.WithGoroutineLabel(ctx, scheduleRegisterRoutineLabel)
	go func(ctx context.Context) {
		pprof.SetGoroutineLabels(ctx)
		defer func() {
			if err := recover(); err != nil {
				w.metrics.ScheduleRegistrationPanicCounter.Inc()
				logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
			}
		}()
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			err := w.ReadAndRegisterSchedules(ctx)
			if err != nil {
				logger.Errorf(ctx, "Failed to iterated over all schedules in the current run due to %v", err)
			}
		}, registrationSleepTime*time.Second)
	}(scheduleRegisterCtxWithLabel)
}

func (w *ScheduleRegisterer) ReadAndRegisterSchedules(ctx context.Context) error {
	schedules, err := w.db.SchedulableEntityRepo().GetAll(ctx)
	if err != nil {
		return fmt.Errorf("unable to run the workflow executor after reading the schedules due to %v", err)
	}

	for _, s := range schedules {
		funcRef := func(jobCtx context.Context, schedule models.SchedulableEntity, scheduleTime time.Time) {
			// If the schedule has been deactivated and then the inflight schedules can stop
			if !*schedule.Active {
				return
			}
			nameOfSchedule := GetScheduleName(ctx, schedule)
			_ = w.rateLimiter.Take()
			err := w.executionFirer.Fire(jobCtx, scheduleTime, schedule)
			if err != nil {
				logger.Errorf(jobCtx, "unable to fire the schedule %+v at %v time due to %v", s, scheduleTime, err)
				return
			} else {
				w.snapshoter.UpdateLastExecutionTime(nameOfSchedule, scheduleTime)
			}
		}

		// Register or deregister the schedule from the scheduler
		if !*s.Active {
			w.goCronInterface.DeRegister(ctx, s)
		} else {
			err := w.goCronInterface.Register(ctx, s, funcRef)
			if err != nil {
				w.metrics.ScheduleRegistrationFailure.Inc()
				logger.Errorf(ctx, "unable to register the schedule %+v due to %v", s, err)
			}
		}
	} // Done iterating over all the read schedules
	return nil
}

func NewScheduleRegisterer(scope promutils.Scope, goCronInterface interfaces.GoCronWrapper,
	snapshoter interfaces.Snapshoter,
	snapShotReaderWriter interfaces.SnapshotReaderWriter,
	db repositories.SchedulerRepoInterface, rateLimiter ratelimit.Limiter,
	executionFirer interfaces.ScheduleExecutionFirer) interfaces.ScheduleRegisterer {

	return &ScheduleRegisterer{
		metrics:              getScheduleRegistererMetrics(scope),
		goCronInterface:      goCronInterface,
		snapshoter:           snapshoter,
		snapShotReaderWriter: snapShotReaderWriter,
		db:                   db,
		rateLimiter:          rateLimiter,
		executionFirer:       executionFirer,
	}
}

func getScheduleRegistererMetrics(scope promutils.Scope) scheduleRegistererMetrics {
	return scheduleRegistererMetrics{
		Scope: scope,
		ScheduleRegistrationPanicCounter: scope.MustNewCounter("schedule_registration_panic_counter",
			"count of crashes for the schedule registration system"),
		ScheduleRegistrationFailure: scope.MustNewCounter("schedule_registration_failure_counter",
			"count of unsuccessful attempts to register the schedules"),
	}
}
