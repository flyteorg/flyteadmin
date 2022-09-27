package gormimpl

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flytestdlib/promutils"

	flyteAdminDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"gorm.io/gorm"
)

// DescriptionEntityRepo Implementation of DescriptionEntityRepoInterface.
type DescriptionEntityRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *DescriptionEntityRepo) Create(ctx context.Context, input models.DescriptionEntity) (uint, error) {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return 0, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	r.db.Last(&input)
	return input.ID, nil
}

func (r *DescriptionEntityRepo) Get(ctx context.Context, input models.DescriptionEntityKey) (models.DescriptionEntity, error) {
	var descriptionEntity models.DescriptionEntity

	filters, err := getDescriptionEntityFilters(input.ResourceType, input.Project, input.Domain, input.Name, input.Version)
	if err != nil {
		return models.DescriptionEntity{}, err
	}

	joinString, joinFound := resourceTypeToDescriptionJoin[input.ResourceType]
	if !joinFound {
		return models.DescriptionEntity{}, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, "Cannot get DescriptionEntity for resource type: %v", input.ResourceType)
	}

	tx := r.db.Table(descriptionEntityTableName).Joins(joinString)

	// Apply filters
	tx, err = applyScopedFilters(tx, filters, nil)
	if err != nil {
		return models.DescriptionEntity{}, err
	}

	timer := r.metrics.GetDuration.Start()
	tx = tx.Take(&descriptionEntity)
	timer.Stop()

	if tx.Error != nil {
		return models.DescriptionEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return descriptionEntity, nil
}

func (r *DescriptionEntityRepo) List(ctx context.Context, input interfaces.ListResourceInput) (
	interfaces.DescriptionEntityCollectionOutput, error) {

	if input.Limit == 0 {
		return interfaces.DescriptionEntityCollectionOutput{}, flyteAdminDbErrors.GetInvalidInputError(limit)
	}

	var descriptionEntities []models.DescriptionEntity
	tx := r.db.Limit(input.Limit).Offset(input.Offset)
	// Apply filters
	tx, err := applyFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.DescriptionEntityCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}
	timer := r.metrics.ListDuration.Start()
	tx.Group(identifierGroupBy).Scan(&descriptionEntities)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.DescriptionEntityCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.DescriptionEntityCollectionOutput{
		Entities: descriptionEntities,
	}, nil
}

var innerJoinDescriptionToTaskName = fmt.Sprintf(
	"INNER JOIN %s ON %s.project = %s.project AND %s.domain = %s.domain AND %s.id = %s.description_id", taskTableName, descriptionEntityTableName, taskTableName,
	descriptionEntityTableName, taskTableName,
	descriptionEntityTableName, taskTableName)

var innerJoinDescriptionToWorkflowName = fmt.Sprintf(
	"INNER JOIN %s ON %s.project = %s.project AND %s.domain = %s.domain AND %s.id = %s.description_id", workflowTableName, descriptionEntityTableName, workflowTableName,
	descriptionEntityTableName, workflowTableName,
	descriptionEntityTableName, workflowTableName)

var innerJoinDescriptionToLaunchPlanName = fmt.Sprintf(
	"INNER JOIN %s ON %s.project = %s.project AND %s.domain = %s.domain AND %s.id = %s.description_id", launchPlanTableName, descriptionEntityTableName, launchPlanTableName,
	descriptionEntityTableName, launchPlanTableName,
	descriptionEntityTableName, launchPlanTableName)

var resourceTypeToDescriptionJoin = map[core.ResourceType]string{
	core.ResourceType_LAUNCH_PLAN: innerJoinDescriptionToLaunchPlanName,
	core.ResourceType_WORKFLOW:    innerJoinDescriptionToWorkflowName,
	core.ResourceType_TASK:        innerJoinDescriptionToTaskName,
}

func getDescriptionEntityFilters(resourceType core.ResourceType, project string, domain string, name string, version string) ([]common.InlineFilter, error) {
	entity := common.ResourceTypeToEntity[resourceType]

	filters := make([]common.InlineFilter, 0)
	projectFilter, err := common.NewSingleValueFilter(entity, common.Equal, Project, project)
	if err != nil {
		return nil, err
	}
	filters = append(filters, projectFilter)
	domainFilter, err := common.NewSingleValueFilter(entity, common.Equal, Domain, domain)
	if err != nil {
		return nil, err
	}
	filters = append(filters, domainFilter)
	nameFilter, err := common.NewSingleValueFilter(entity, common.Equal, Name, name)
	if err != nil {
		return nil, err
	}
	filters = append(filters, nameFilter)
	versionFilter, err := common.NewSingleValueFilter(entity, common.Equal, Version, version)
	if err != nil {
		return nil, err
	}
	filters = append(filters, versionFilter)

	return filters, nil
}

// NewDescriptionEntityRepo Returns an instance of DescriptionRepoInterface
func NewDescriptionEntityRepo(
	db *gorm.DB, errorTransformer flyteAdminDbErrors.ErrorTransformer, scope promutils.Scope) interfaces.DescriptionEntityRepoInterface {
	metrics := newMetrics(scope)
	return &DescriptionEntityRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
