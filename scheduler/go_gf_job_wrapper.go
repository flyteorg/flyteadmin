package scheduler

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/os/gcron"
	"github.com/gogf/gf/os/gtimer"
	"time"
)

type GoGfJobWrapper struct {
	ctx                context.Context
	nameOfSchedule     string
	schedule           models.SchedulableEntity
	AsOfTime           time.Time
	TempAsOfTime       time.Time
	fixedIntervalEntry *gtimer.Entry
	jobFunc            func(ticks int64)
}

func (g *GoGfJobWrapper) ScheduleJob() {
	s := g.schedule
	if len(g.schedule.CronExpression) > 0 {
		err := g.AddCronJob()
		if err != nil {
			logger.Errorf(g.ctx, "failed to add cron schedule %v due to %v", s, err)
		}
	} else {
		err := g.AddFixedIntervalJob()
		if err != nil {
			logger.Errorf(g.ctx, "failed to add fixed rate schedule %v due to %v", s, err)
		}
	}
}

func (g *GoGfJobWrapper) DeScheduleJob() {
	s := g.schedule
	if len(s.CronExpression) > 0 {
		g.RemoveCronJob()
	} else {
		g.RemoveFixedIntervalJob()
	}
}

func (g *GoGfJobWrapper) AddFixedIntervalJob() error {
	if g.fixedIntervalEntry != nil {
		// Already exists
		return nil
	}
	d, err := getFixedRateDurationFromSchedule(g.schedule.Unit, g.schedule.FixedRateValue)
	if err != nil {
		return err
	}

	logger.Infof(g.ctx, "successfully added the fixed rate schedule %s to the scheduler", g.nameOfSchedule)

	g.fixedIntervalEntry = gtimer.AddTimedJob(d, g.jobFunc)
	return nil
}

func (g *GoGfJobWrapper) RemoveFixedIntervalJob() {
	// TODO : find the right way to remove the fixed interval job
	if g.fixedIntervalEntry == nil {
		// Entry doesn't exist
		return
	}
	g.fixedIntervalEntry.Stop()
	logger.Infof(g.ctx, "successfully removed the schedule %s from scheduler", g.nameOfSchedule)
}

func (g *GoGfJobWrapper) AddCronJob() error {
	_, err := gcron.AddTimedJob(g.schedule.CronExpression, g.jobFunc, g.nameOfSchedule)
	if err != nil && gerror.Code(err) == gerror.CodeInvalidOperation {
		return nil
	}
	if err == nil {
		logger.Infof(g.ctx, "successfully added the schedule %s to the scheduler", g.nameOfSchedule)
	}
	return err
}

func (g *GoGfJobWrapper) RemoveCronJob() {
	if e := gcron.Search(g.nameOfSchedule); e != nil {
		gcron.Remove(g.nameOfSchedule)
		logger.Infof(g.ctx, "successfully removed the schedule %s from scheduler", g.nameOfSchedule)
	}
}
