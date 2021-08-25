package executor

import (
	"context"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/robfig/cron/v3"
)

type GoCronJobWrapper struct {
	ctx            context.Context
	c              *cron.Cron
	nameOfSchedule string
	entryId        cron.EntryID
	schedule       models.SchedulableEntity
	jobFunc        cron.TimedFuncJob
}

func (g *GoCronJobWrapper) ScheduleJob() {
	s := g.schedule
	if len(g.schedule.CronExpression) > 0 {
		err := g.AddCronJob()
		if err != nil {
			logger.Errorf(g.ctx, "failed to add cron schedule %+v due to %v", s, err)
		}
	} else {
		err := g.AddFixedIntervalJob()
		if err != nil {
			logger.Errorf(g.ctx, "failed to add fixed rate schedule %+v due to %v", s, err)
		}
	}
}

func (g *GoCronJobWrapper) DeScheduleJob() {
	s := g.schedule
	if len(s.CronExpression) > 0 {
		g.RemoveCronJob()
	} else {
		g.RemoveFixedIntervalJob()
	}
}

func (g *GoCronJobWrapper) AddFixedIntervalJob() error {
	d, err := getFixedRateDurationFromSchedule(g.schedule.Unit, g.schedule.FixedRateValue)
	if err != nil {
		return err
	}

	g.c.ScheduleTimedJob(cron.ConstantDelaySchedule{Delay: d}, g.jobFunc)
	logger.Infof(g.ctx, "successfully added the fixed rate schedule %s to the scheduler for schedule %+v",
		g.nameOfSchedule, g.schedule)

	return nil
}

func (g *GoCronJobWrapper) RemoveFixedIntervalJob() {
	g.c.Remove(g.entryId)
	logger.Infof(g.ctx, "successfully removed the schedule %s from scheduler for schedule %+v",
		g.nameOfSchedule, g.schedule)
}

func (g *GoCronJobWrapper) AddCronJob() error {
	entryId, err := g.c.AddTimedJob(g.schedule.CronExpression, g.jobFunc)
	g.entryId = entryId
	if err == nil {
		logger.Infof(g.ctx, "successfully added the schedule %s to the scheduler for schedule %+v",
			g.nameOfSchedule, g.schedule)
	}
	return err
}

func (g *GoCronJobWrapper) RemoveCronJob() {
	g.c.Remove(g.entryId)
	logger.Infof(g.ctx, "successfully removed the schedule %s from scheduler for schedue %+v",
		g.nameOfSchedule, g.schedule)

}
