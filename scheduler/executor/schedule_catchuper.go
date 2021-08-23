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
	"runtime/debug"
	"runtime/pprof"
	"time"
)

type catchuperMetrics struct {
	Scope               promutils.Scope
	CatchupPanicCounter prometheus.Counter
	CatchupErrCounter   prometheus.Counter
}

type ScheduleCatchuper struct {
	metrics         catchuperMetrics
	snapshoter      interfaces.Snapshoter
	goCronInterface interfaces.GoCronWrapper
	rateLimiter     ratelimit.Limiter
	db              repositories.SchedulerRepoInterface
	executionFirer  interfaces.ScheduleExecutionFirer
}

func (c *ScheduleCatchuper) RunCatchuper(ctx context.Context) chan error {
	catchUpContextWithLabel := contextutils.WithGoroutineLabel(ctx, catchupRoutineLabel)
	catchUpErrChan := make(chan error, 1)
	go func(ctx context.Context) {
		pprof.SetGoroutineLabels(ctx)
		defer func() {
			if err := recover(); err != nil {
				c.metrics.CatchupPanicCounter.Inc()
				logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
			}
		}()
		catchUpErrChan <- c.ReadAndCatchupSchedules(ctx)
	}(catchUpContextWithLabel)

	return catchUpErrChan
}

func (w *ScheduleCatchuper) ReadAndCatchupSchedules(ctx context.Context) error {
	schedules, err := w.db.SchedulableEntityRepo().GetAll(ctx)
	if err != nil {
		logger.Errorf(ctx, "unable to read the schedules due to %v", err)
		return err
	}
	// Run the catchup system
	catchUpTill := time.Now()
	err = w.CatchUpAllSchedules(ctx, schedules, catchUpTill)
	if err != nil {
		logger.Errorf(ctx, "unable to catch up on the schedules due to %v", err)
		return err
	}
	return nil
}

func (w *ScheduleCatchuper) CatchUpAllSchedules(ctx context.Context, schedules []models.SchedulableEntity, toTime time.Time) error {
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
		nameOfSchedule := GetScheduleName(ctx, s)
		lastT := w.snapshoter.GetLastExecutionTime(nameOfSchedule)
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

func (w *ScheduleCatchuper) CatchUpSingleSchedule(ctx context.Context, s models.SchedulableEntity, fromTime time.Time, toTime time.Time) error {
	nameOfSchedule := GetScheduleName(ctx, s)
	var catchUpTimes []time.Time
	var err error
	catchUpTimes, err = w.goCronInterface.GetCatchUpTimes(s, fromTime, toTime)
	if err != nil {
		return err
	}
	var catchupTime time.Time
	for _, catchupTime = range catchUpTimes {
		_ = w.rateLimiter.Take()
		err := w.executionFirer.Fire(ctx, catchupTime, s)
		if err != nil {
			w.metrics.CatchupErrCounter.Inc()
			logger.Errorf(ctx, "unable to fire the schedule %+v at %v time due to %v", s, catchupTime, err)
			return err
		} else {
			w.snapshoter.UpdateLastExecutionTime(nameOfSchedule, catchupTime)
		}
	}

	return nil
}

func NewScheduleCatchuper(scope promutils.Scope, snapshoter interfaces.Snapshoter,
	goCronInterface interfaces.GoCronWrapper, rateLimiter ratelimit.Limiter,
	db repositories.SchedulerRepoInterface,
	executionFirer interfaces.ScheduleExecutionFirer) interfaces.ScheduleCatchuper {

	return &ScheduleCatchuper{
		metrics:         getCatchupMetrics(scope),
		snapshoter:      snapshoter,
		goCronInterface: goCronInterface,
		rateLimiter:     rateLimiter,
		db:              db,
		executionFirer:  executionFirer,
	}
}

func getCatchupMetrics(scope promutils.Scope) catchuperMetrics {
	return catchuperMetrics{
		Scope: scope,
		CatchupPanicCounter: scope.MustNewCounter("catchup_panic_counter",
			"count of crashes for the catchup system"),
		CatchupErrCounter: scope.MustNewCounter("catchup_error_counter",
			"count of unsuccessful attempts to catchup on the schedules"),
	}
}
