package scheduler

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	schedinterfaces "github.com/flyteorg/flyteadmin/scheduler/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/robfig/cron"
	"time"
)

// GoCron Struct implementing the GoCronWrapper which is used by the scheduler for adding and removing schedules
// Each scheduled job accepts scheduled time parameter which helps to know what the actual invocation time and use
// that to send an execution to admin
type GoCron struct {
	jobsMap map[string]schedinterfaces.GoCronJobWrapper
	c       *cron.Cron
}

func (g GoCron) DeRegister(ctx context.Context, s models.SchedulableEntity) {
	nameOfSchedule := GetScheduleName(s)

	if g.jobsMap[nameOfSchedule] == nil {
		logger.Debugf(ctx, "Job doesn't exists in the map for name %v with schedule %+v  and hence already removed", nameOfSchedule, s)
		return
	}
	g.jobsMap[nameOfSchedule].DeScheduleJob()

	// Delete it from the jobs map
	delete(g.jobsMap, nameOfSchedule)
}

func (g GoCron) Register(ctx context.Context, s models.SchedulableEntity, registerFuncRef schedinterfaces.RegisterFuncRef) error {
	nameOfSchedule := GetScheduleName(s)

	if g.jobsMap[nameOfSchedule] != nil {
		logger.Debugf(ctx, "Job already exists in the map for name %v with schedule %+v", nameOfSchedule, s)
		return nil
	}

	job := &GoCronJobWrapper{schedule: s, ctx: ctx, nameOfSchedule: nameOfSchedule, c: g.c}
	g.jobsMap[nameOfSchedule] = job

	job.jobFunc = func(triggerTime time.Time) {
		registerFuncRef(ctx, s, triggerTime)
	}
	job.ScheduleJob()
	return nil
}

func (g GoCron) GetScheduledTime(s models.SchedulableEntity, fromTime time.Time) (time.Time, error) {
	if len(s.CronExpression) > 0 {
		return getCronScheduledTime(s.CronExpression, fromTime)
	} else {
		return getFixedIntervalScheduledTime(s.Unit, s.FixedRateValue, fromTime)
	}
}

func (g GoCron) GetCatchUpTimes(schedule models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error) {
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

func getCronScheduledTime(cronString string, fromTime time.Time) (time.Time, error) {
	sched, err := cron.ParseStandard(cronString)
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
