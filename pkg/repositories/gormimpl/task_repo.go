package gormimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/promutils"

	flyteAdminDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"gorm.io/gorm"
)

// Implementation of TaskRepoInterface.
type TaskRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *TaskRepo) Create(ctx context.Context, input models.Task) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *TaskRepo) UpdateDescriptionID(input models.Task) error {
	timer := r.metrics.UpdateDuration.Start()
	tx := r.db.Where(&models.Task{
		TaskKey: models.TaskKey{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
			Version: input.Version,
		},
	}).Assign(input).FirstOrCreate(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *TaskRepo) Get(ctx context.Context, input interfaces.Identifier) (models.Task, error) {
	var task models.Task
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.Task{
		TaskKey: models.TaskKey{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
			Version: input.Version,
		},
	})
	tx = tx.Joins(leftJoinTaskToDescription)
	tx = tx.Select([]string{
		fmt.Sprintf("%s.*", taskTableName),
		fmt.Sprintf("%s.short_description", descriptionEntityTableName),
	}).Take(&task)
	timer.Stop()
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.Task{}, flyteAdminDbErrors.GetMissingEntityError(core.ResourceType_TASK.String(), &core.Identifier{
			Project: input.Project,
			Domain:  input.Domain,
			Name:    input.Name,
			Version: input.Version,
		})
	}

	if tx.Error != nil {
		return models.Task{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return task, nil
}

func (r *TaskRepo) List(
	ctx context.Context, input interfaces.ListResourceInput) (interfaces.TaskCollectionOutput, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.TaskCollectionOutput{}, err
	}
	var tasks []models.Task
	tx := r.db.Limit(input.Limit).Offset(input.Offset)
	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.TaskCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}
	timer := r.metrics.ListDuration.Start()
	tx = tx.Joins(leftJoinTaskToDescription)
	tx = tx.Select([]string{
		fmt.Sprintf("%s.*", taskTableName),
		fmt.Sprintf("%s.short_description", descriptionEntityTableName),
	}).Find(&tasks)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.TaskCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.TaskCollectionOutput{
		Tasks: tasks,
	}, nil
}

func (r *TaskRepo) ListTaskIdentifiers(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.TaskCollectionOutput, error) {

	// Validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.TaskCollectionOutput{}, err
	}

	tx := r.db.Model(models.Task{}).Limit(input.Limit).Offset(input.Offset)

	// Apply filters
	tx, err := applyFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.TaskCollectionOutput{}, err
	}
	for _, mapFilter := range input.MapFilters {
		tx = tx.Where(mapFilter.GetFilter())
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	// Scan the results into a list of tasks
	var tasks []models.Task
	timer := r.metrics.ListIdentifiersDuration.Start()
	tx.Select([]string{Project, Domain, Name}).Group(identifierGroupBy).Scan(&tasks)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.TaskCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.TaskCollectionOutput{
		Tasks: tasks,
	}, nil
}

var leftJoinTaskToDescription = fmt.Sprintf(
	"LEFT JOIN %s ON %s.project = %s.project AND %s.domain = %s.domain AND %s.id = %s.description_id", descriptionEntityTableName, descriptionEntityTableName, taskTableName,
	descriptionEntityTableName, taskTableName,
	descriptionEntityTableName, taskTableName)

// Returns an instance of TaskRepoInterface
func NewTaskRepo(
	db *gorm.DB, errorTransformer flyteAdminDbErrors.ErrorTransformer, scope promutils.Scope) interfaces.TaskRepoInterface {
	metrics := newMetrics(scope)
	return &TaskRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
