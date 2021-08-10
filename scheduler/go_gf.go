package scheduler

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	schedinterfaces "github.com/flyteorg/flyteadmin/scheduler/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/os/gcron"
	"github.com/gogf/gf/os/gtimer"
	"github.com/robfig/cron"
	"time"
)

// GoGF Struct implementing the GoGFWrapper which is used by the scheduler for adding and removing schedules
type GoGF struct {
	fixedIntervalEntries map[string]*gtimer.Entry
}

func (g GoGF) Register(ctx context.Context, s models.SchedulableEntity, registerFuncRef schedinterfaces.RegisterFuncRef) error {
	nameOfSchedule := GetScheduleName(s)

	jobFunc := func(ticks int64) {
		d := time.Duration(ticks) * time.Millisecond * 100
		fromTime := s.UpdatedAt.Add(d)
		scheduledTime, err := g.GetScheduledTime(s, fromTime)
		if err != nil {
			logger.Errorf(ctx, "unable to get scheduled time from cron expression for %+v schedule due to %v", s, err)
			return
		}
		registerFuncRef(ctx, s, scheduledTime)
	}
	// Register activation record by adding a new schedule if it doesn't exist
	if *s.Active {
		if len(s.CronExpression) > 0 {
			err :=g.addCronJob(ctx, s.CronExpression, jobFunc, nameOfSchedule)
			if err != nil {
				logger.Errorf(ctx, "failed to add cron schedule %v due to %v", s, err)
			}
		} else {
			err := g.addFixedIntervalJob(ctx, s.Unit, s.FixedRateValue, jobFunc, nameOfSchedule)
			if err != nil {
				logger.Errorf(ctx, "failed to add fixed rate schedule %v due to %v", s, err)
			}
		}
	} else {
		// Register deactivation record by removing the schedule if it exists
		if len(s.CronExpression) > 0 {
			g.removeCronJob(ctx, nameOfSchedule)
		} else {
			g.removeFixedIntervalJob(ctx, nameOfSchedule)
		}
	}
	return nil
}

func (g GoGF) GetScheduledTime(s models.SchedulableEntity, fromTime time.Time) (time.Time, error) {
	if len(s.CronExpression) > 0 {
		return getCronScheduledTime(s.CronExpression, fromTime)
	} else {
		return getFixedIntervalScheduledTime(s.Unit, s.FixedRateValue, fromTime)
	}
}

func (g GoGF) GetCatchUpTimes(schedule models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error) {
	var scheduledTimes []time.Time
	currFrom := from
	for currFrom.Before(to) {
		scheduledTime, err := g.GetScheduledTime(schedule, currFrom)
		if err != nil {
			return nil, err
		}
		scheduledTimes = append(scheduledTimes, scheduledTime)
		currFrom = scheduledTime
	}
	return scheduledTimes, nil
}

func (g GoGF) addCronJob(ctx context.Context, cronExpression string, job func(int64), nameOfSchedule string) error {
	_, err := gcron.AddTimedJob(cronExpression, job, nameOfSchedule)
	if err != nil && gerror.Code(err) == gerror.CodeInvalidOperation {
		return nil
	}
	if err == nil {
		logger.Infof(ctx, "successfully added the schedule %s to the scheduler", nameOfSchedule)
	}
	return err
}

func (g GoGF) removeCronJob(ctx context.Context, nameOfSchedule string) {
	if e := gcron.Search(nameOfSchedule); e != nil {
		gcron.Remove(nameOfSchedule)
		logger.Infof(ctx, "successfully removed the schedule %s from scheduler", nameOfSchedule)
	}
}

func (g GoGF) addFixedIntervalJob(ctx context.Context, unit admin.FixedRateUnit, fixedRateValue uint32, job func(int64), nameOfSchedule string) error {
	if g.fixedIntervalEntries[nameOfSchedule] != nil {
		// Already exists
		return nil
	}
	d, err := getFixedRateDurationFromSchedule(unit, fixedRateValue)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "successfully added the fixed rate schedule %s to the scheduler", nameOfSchedule)

	g.fixedIntervalEntries[nameOfSchedule] = gtimer.AddTimedJob(d, job)
	return nil
}

func (g GoGF) removeFixedIntervalJob(ctx context.Context, nameOfSchedule string) {
	// TODO : find the right way to remove the fixed interval job
	if g.fixedIntervalEntries[nameOfSchedule] == nil {
		// Entry doesn't exist
		return
	}
	g.fixedIntervalEntries[nameOfSchedule].Stop()
	delete(g.fixedIntervalEntries, nameOfSchedule)
	logger.Infof(ctx, "successfully removed the schedule %s from scheduler", nameOfSchedule)
}

func getCronScheduledTime(cronString string, fromTime time.Time) (time.Time, error) {
	var secondParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	sched, err := secondParser.Parse(cronString)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(fromTime), nil
}

func getFixedIntervalScheduledTime(unit admin.FixedRateUnit, fixedRateValue uint32, fromTime time.Time) (time.Time, error) {
	d, err := getFixedRateDurationFromSchedule(unit, fixedRateValue)
	if err != nil {
		return time.Time{}, err
	}
	fixedRateSchedule := cron.ConstantDelaySchedule{Delay: d}
	return fixedRateSchedule.Next(fromTime), nil
}

func getFixedRateDurationFromSchedule(unit admin.FixedRateUnit, fixedRateValue uint32) (time.Duration, error) {
	d := time.Duration(fixedRateValue)
	switch unit {
	case admin.FixedRateUnit_MINUTE:
		d = d * time.Minute
	case admin.FixedRateUnit_HOUR:
		d = d * time.Hour
	case admin.FixedRateUnit_DAY:
		d = d * time.Hour * 24
	default:
		return -1, fmt.Errorf("unsupported unit %v for fixed rate scheduling ", unit)
	}
	return d, nil
}
