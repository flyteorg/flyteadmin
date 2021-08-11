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

// GoGF Struct implementing the GoGFWrapper which is used by the scheduler for adding and removing schedules
type GoGF struct {
	jobsMap map[string]schedinterfaces.GoGFJobWrapper
}

func (g GoGF) DeRegister(ctx context.Context, s models.SchedulableEntity) {
	nameOfSchedule := GetScheduleName(s)

	if g.jobsMap[nameOfSchedule] == nil {
		logger.Debugf(ctx, "Job doesn't exists in the map for name %v with schedule %+v  and hence already removed", nameOfSchedule, s)
		return
	}
	g.jobsMap[nameOfSchedule].DeScheduleJob()

	// Delete it from the jobs map
	delete(g.jobsMap, nameOfSchedule)
}

func (g GoGF) Register(ctx context.Context, s models.SchedulableEntity, asOfTime time.Time, registerFuncRef schedinterfaces.RegisterFuncRef) error {
	nameOfSchedule := GetScheduleName(s)

	if g.jobsMap[nameOfSchedule] != nil {
		logger.Debugf(ctx, "Job already exists in the map for name %v with schedule %+v", nameOfSchedule, s)
		return nil
	}

	// Set the temporary as of time before the first callback
	// First call back will correctly set the AsOfTime
	tempAsOfTime := s.UpdatedAt
	if !asOfTime.IsZero() && asOfTime.After(s.UpdatedAt) {
		tempAsOfTime = asOfTime
	}

	job := &GoGfJobWrapper{schedule: s, TempAsOfTime: tempAsOfTime, ctx: ctx, nameOfSchedule: nameOfSchedule}
	g.jobsMap[nameOfSchedule] = job

	job.jobFunc = func(ticks int64) {
		lockedScheduleTime, err := g.GetScheduledTime(s, job.TempAsOfTime)
		if err != nil {
			logger.Errorf(ctx, "unable to get next scheduled time for %+v schedule due to %v", s, err)
			return
		}
		job.TempAsOfTime = lockedScheduleTime
		//var err error
		//if job.AsOfTime.IsZero() {
		//	lockedScheduleTime, err = g.GetScheduledTime(s, job.TempAsOfTime)
		//	if err != nil {
		//		logger.Errorf(ctx, "unable to get next scheduled time for %+v schedule due to %v", s, err)
		//		return
		//	}
		//	glog.Printf("Going to fire at %v", job.TempAsOfTime)
		//	d := time.Duration(ticks) * time.Millisecond * 100
		//	job.AsOfTime = time.Now().Add(-d)
		//} else {
		//	d := time.Duration(ticks) * time.Millisecond * 100
		//	glog.Printf("Going to fire at %v", job.AsOfTime.Add(d))
		//	lockedScheduleTime = job.AsOfTime.Add(d)
		//}
		registerFuncRef(ctx, s, lockedScheduleTime)
	}
	job.ScheduleJob()

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
