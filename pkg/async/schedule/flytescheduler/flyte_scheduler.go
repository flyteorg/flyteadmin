package flytescheduler

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	scheduleInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

type FlyteScheduler struct {
	db repositories.RepositoryInterface
}

func (s *FlyteScheduler) CreateScheduleInput(ctx context.Context, appConfig *runtimeInterfaces.SchedulerConfig,
	identifier core.Identifier, schedule *admin.Schedule) (interfaces.AddScheduleInput, error) {

	addScheduleInput := scheduleInterfaces.AddScheduleInput{
		Identifier:         identifier,
		ScheduleExpression: *schedule,
	}
	return addScheduleInput, nil
}

func (s *FlyteScheduler) AddSchedule(ctx context.Context, input interfaces.AddScheduleInput) error {
	logger.Debugf(ctx, "Received call to add schedule [%+v]", input)
	var cronString string
	switch v := input.ScheduleExpression.GetScheduleExpression().(type) {
	case *admin.Schedule_Rate:
		cronString = fmt.Sprintf("*/%v * * * * *", v.Rate.Value)
	case *admin.Schedule_CronSchedule:
		cronString = v.CronSchedule.Schedule
	case *admin.Schedule_CronExpression:
		cronString = v.CronExpression
	default:
		return fmt.Errorf("failed adding schedule for unknown schedule expression type %v", v)
	}
	modelInput := models.SchedulableEntity{
		CronExpression: cronString,
		KickoffTimeInputArg: input.ScheduleExpression.KickoffTimeInputArg,
		Active: true,
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: input.Identifier.Project,
			Domain: input.Identifier.Domain,
			Name: input.Identifier.Name,
			Version: input.Identifier.Version,
		},
		//BaseModel: models.BaseModel{
		//
		//},
	}
	err := s.db.SchedulableEntityRepo().Create(ctx, modelInput)
	if err != nil {
		return err
	}
	logger.Debug(ctx, "Created a scheduled entity for %v ", input)
	return nil
}

func (s *FlyteScheduler) RemoveSchedule(ctx context.Context, input interfaces.RemoveScheduleInput) error {
	logger.Debugf(ctx, "Received call to remove schedule [%+v]", input.Identifier)

	//schedulableEntity, err := s.db.SchedulableEntityRepo().Get(ctx, models.SchedulableEntityKey{
	//	Project: input.Identifier.Project,
	//	Domain: input.Identifier.Domain,
	//	Name: input.Identifier.Name,
	//	Version: input.Identifier.Version,
	//})
	//if err != nil {
	//	return err
	//}
	logger.Debug(ctx, "Not scheduling anything")
	return nil
}

func NewFlyteScheduler(db repositories.RepositoryInterface) interfaces.EventScheduler {
	return &FlyteScheduler{db: db}
}
