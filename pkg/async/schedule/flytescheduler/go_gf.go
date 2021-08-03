package flytescheduler

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/aws"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/gogf/gf/os/gcron"
	"github.com/gogf/gf/os/gtimer"
	"github.com/robfig/cron"
	"strconv"
	"strings"
	"time"
)

type GoGF struct {
}

func (g GoGF) GetScheduleName(s models.SchedulableEntity) string {
	return strconv.FormatUint(aws.HashIdentifier(core.Identifier{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    s.Name,
		Version: s.Version,
	}), 10)
}

func (g GoGF) Register(ctx context.Context, s models.SchedulableEntity, funcRef func()) error {
	nameOfSchedule := g.GetScheduleName(s)

	if s.Active {
		if len(s.CronExpression) > 0 {
			err := addCronJob(s.CronExpression,funcRef,nameOfSchedule)
			if err != nil {
				logger.Errorf(ctx, "failed to add cron schedule %v due to %v", s, err)
			}
		} else {
			err := addFixedIntervalJob(s.Unit, s.FixedRateValue, funcRef)
			if err != nil {
				logger.Errorf(ctx, "failed to add fixed rate schedule %v due to %v", s, err)
			}
		}
	} else {
		if len(s.CronExpression) > 0 {
			removeCronJob(ctx, nameOfSchedule)
		} else {
			removeFixedIntervalJob(ctx, nameOfSchedule)
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

func (g GoGF) GetCatchUpTimes(schedule models.SchedulableEntity, from time.Time, to time.Time) ([]time.Time, error){
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

func addCronJob(cronExpression string, job func(), nameOfSchedule string) error {
	_, err := gcron.AddSingleton(cronExpression, job, nameOfSchedule)
	if err != nil && strings.HasSuffix(err.Error(), "already exists"){
		return nil
	}
	return err
}

func removeCronJob(ctx context.Context, nameOfSchedule string) {
	gcron.Remove(nameOfSchedule)
	logger.Infof(ctx, "successfully removed the schedule %v from scheduler", nameOfSchedule)
}

func addFixedIntervalJob(unit admin.FixedRateUnit, fixedRateValue uint32, job func()) error {
	d, err := getFixedRateDurationFromSchedule(unit, fixedRateValue)
	if err != nil {
		return err
	}
	gtimer.AddSingleton(d, job)
	return nil
}

func removeFixedIntervalJob(ctx context.Context, nameOfSchedule string) {
	// TODO : find the right way to remove the fixed interval job
	logger.Infof(ctx, "successfully remove the schedule %v from scheduler", nameOfSchedule)
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
		d = d * time.Second
	case admin.FixedRateUnit_HOUR:
		d = d * time.Hour
	case admin.FixedRateUnit_DAY:
		d = d * time.Hour * 24
	default:
		return -1, fmt.Errorf("unsupported unit %v for fixed rate scheduling ", unit)
	}
	return d, nil
}