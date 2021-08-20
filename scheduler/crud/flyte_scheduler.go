package crud

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

// FlyteScheduler used for saving the scheduler entries after launch plans are enabled or disabled.
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
	logger.Infof(ctx, "Received call to add schedule [%+v]", input)
	var cronString string
	var fixedRateValue uint32
	var fixedRateUnit admin.FixedRateUnit
	switch v := input.ScheduleExpression.GetScheduleExpression().(type) {
	case *admin.Schedule_Rate:
		fixedRateValue = v.Rate.Value
		fixedRateUnit = v.Rate.Unit
	case *admin.Schedule_CronSchedule:
		cronString = v.CronSchedule.Schedule
	case *admin.Schedule_CronExpression:
		cronString = v.CronExpression
	default:
		return fmt.Errorf("failed adding schedule for unknown schedule expression type %v", v)
	}
	active := true
	modelInput := models.SchedulableEntity{
		CronExpression:      cronString,
		FixedRateValue:      fixedRateValue,
		Unit:                fixedRateUnit,
		KickoffTimeInputArg: input.ScheduleExpression.KickoffTimeInputArg,
		Active:              &active,
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: input.Identifier.Project,
			Domain:  input.Identifier.Domain,
			Name:    input.Identifier.Name,
			Version: input.Identifier.Version,
		},
	}
	err := s.db.SchedulableEntityRepo().Activate(ctx, modelInput)
	if err != nil {
		return err
	}
	logger.Infof(ctx, "Activated scheduled entity for %v ", input)
	return nil
}

func (s *FlyteScheduler) RemoveSchedule(ctx context.Context, input interfaces.RemoveScheduleInput) error {
	logger.Infof(ctx, "Received call to remove schedule [%+v]. Will deactivate it in the scheduler", input.Identifier)

	err := s.db.SchedulableEntityRepo().Deactivate(ctx, models.SchedulableEntityKey{
		Project: input.Identifier.Project,
		Domain:  input.Identifier.Domain,
		Name:    input.Identifier.Name,
		Version: input.Identifier.Version,
	})

	if err != nil {
		return err
	}
	logger.Infof(ctx, "Deactivated the schedule %v in the scheduler", input)
	return nil
}

func NewFlyteScheduler(db repositories.RepositoryInterface) interfaces.EventScheduler {
	return &FlyteScheduler{db: db}
}
