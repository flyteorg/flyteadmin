package server

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"time"
)

func getDb(ctx context.Context) (interfaces.Repository, error) {
	configuration := runtime.NewConfigurationProvider()
	databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
	logConfig := logger.GetConfig()
	scope := promutils.NewScope(configuration.ApplicationConfiguration().GetTopLevelConfig().GetMetricsScope()).NewSubScope("cleanup")

	db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
	if err != nil {
		return nil, err
	}

	repos := repositories.NewGormRepo(db, errors.NewPostgresErrorTransformer(scope.NewSubScope("errors")), scope)
	// TODO: Verify if we need nil check in repos
	return repos, nil
}

func Cleanup(ctx context.Context, retention int) error {
	now := time.Now()
	dateFilter := now.AddDate(0, 0, -retention)
	db, err := getDb(ctx)
	if err != nil {
		return err
	}

	cleanExecution(ctx, db, dateFilter)
	cleanNodeExecution(ctx, db, dateFilter)
	cleanNodeExecutionEvent(ctx, db, dateFilter)
	cleanTaskExecution(ctx, db, dateFilter)

	return nil
}

func cleanExecution(ctx context.Context, db interfaces.Repository, filter time.Time) (int, error) {
	executionList, err := db.ExecutionRepo().List(ctx, createFilter(filter))
	if err != nil {
		return 0, err
	}
	// Do we need counter?
	counter := 0
	for _, e := range executionList.Executions {
		// Check for Gorm Documentation to see how Delete works (Batching etc...)
		err := db.ExecutionRepo().Delete(ctx, e)
		if err != nil {
			return 0, err
		}
		counter++
	}
	return counter, nil
}

func cleanNodeExecution(ctx context.Context, db interfaces.Repository, filter time.Time) (int, error) {
	nodeExecutionList, err := db.NodeExecutionRepo().List(ctx, createFilter(filter))
	if err != nil {
		return 0, err
	}
	counter := 0
	for _, e := range nodeExecutionList.NodeExecutions {
		err := db.NodeExecutionRepo().Delete(ctx, e)
		if err != nil {
			return 0, err
		}
		counter++
	}
	return counter, nil
}

func cleanNodeExecutionEvent(ctx context.Context, db interfaces.Repository, filter time.Time) (int, error) {
	nodeExecutionEventList, err := db.NodeExecutionEventRepo().List(ctx, createFilter(filter))
	if err != nil {
		return 0, err
	}
	counter := 0
	for _, e := range nodeExecutionEventList.NodeExecutionEvents {
		err := db.NodeExecutionEventRepo().Delete(ctx, e)
		if err != nil {
			return 0, err
		}
		counter++
	}
	return counter, nil
}

func cleanTaskExecution(ctx context.Context, db interfaces.Repository, filter time.Time) (int, error) {
	taskList, err := db.TaskExecutionRepo().List(ctx, createFilter(filter))
	if err != nil {
		return 0, err
	}
	counter := 0
	for _, e := range taskList.TaskExecutions {
		err := db.TaskExecutionRepo().Delete(ctx, e)
		if err != nil {
			return 0, err
		}
		counter++
	}
	return counter, nil
}

func createFilter(filter time.Time) interfaces.ListResourceInput {
	return interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			createInlineFilter(common.Execution, "created_at", filter),
		},
		// We need to set a limit, need to check if there's pagination available
		Limit: 1000,
	}
}

func createInlineFilter(entity common.Entity, field string, value interface{}) common.InlineFilter {
	filter, _ := common.NewSingleValueFilter(entity, common.LessThan, field, value)
	return filter
}
