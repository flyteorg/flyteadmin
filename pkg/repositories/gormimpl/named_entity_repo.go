package gormimpl

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"

	"github.com/jinzhu/gorm"
	adminErrors "github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flytestdlib/promutils"
)

// Implementation of NamedEntityRepoInterface.
type NamedEntityRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *NamedEntityRepo) Update(ctx context.Context, input models.NamedEntity) error {
	timer := r.metrics.UpdateDuration.Start()
	var metadata models.NamedEntityMetadata
	tx := r.db.Debug().Where(&models.NamedEntityMetadata{
		NamedEntityMetadataKey: models.NamedEntityMetadataKey{
			ResourceType: input.ResourceType,
			Project:      input.Project,
			Domain:       input.Domain,
			Name:         input.Name,
		},
	}).Assign(input.NamedEntityMetadataFields).FirstOrCreate(&metadata)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *NamedEntityRepo) Get(ctx context.Context, input interfaces.GetNamedEntityInput) (models.NamedEntity, error) {
	var namedEntity models.NamedEntity

	// This is implemented as a filter+join to ensure that when the provided
	// resource_type/project/domain/name combination don't map to a valid
	// entry in the target table, we return NotFound instead of a fake record
	// with empty metadata.
	filters, err := getNamedEntityFilters(input.ResourceType, input.Project, input.Domain, input.Name)
	if err != nil {
		return models.NamedEntity{}, err
	}

	tableName, tableFound := resourceTypeToTableName[resourceType]
	joinString, joinFound := resourceTypeToMetadataJoin[resourceType]
	if !tableFound || !joinFound {
		return models.NamedEntity{}, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, "Cannot get NamedEntity for resource type: %v", resourceType)
	}

	tx := r.db.Table(tableName).Joins(joinString)

	// Apply filters
	tx, err = applyScopedFilters(tx, filters, nil)
	if err != nil {
		return models.NamedEntity{}, err
	}

	timer := r.metrics.GetDuration.Start()
	tx = tx.Select(getSelectForNamedEntity(tableName)).First(&namedEntity)
	timer.Stop()

	if tx.Error != nil {
		return models.NamedEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return namedEntity, nil
}

func (r *NamedEntityRepo) List(ctx context.Context, resourceType core.ResourceType, input interfaces.ListResourceInput) (
	interfaces.NamedEntityCollectionOutput, error) {

	// Validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.NamedEntityCollectionOutput{}, err
	}

	tableName, tableFound := resourceTypeToTableName[resourceType]
	joinString, joinFound := resourceTypeToMetadataJoin[resourceType]
	if !tableFound || !joinFound {
		return interfaces.NamedEntityCollectionOutput{}, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, "Cannot list entity names for resource type: %v", resourceType)
	}

	tx := r.db.Table(tableName).Limit(input.Limit).Offset(input.Offset)
	tx = tx.Joins(joinString)

	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.NamedEntityCollectionOutput{}, err
	}

	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	// Scan the results into a list of named entities
	var entities []models.NamedEntity
	timer := r.metrics.ListDuration.Start()
	tx.Select(getSelectForNamedEntity(tableName)).Group(getGroupByForNamedEntity(tableName)).Scan(&entities)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.NamedEntityCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.NamedEntityCollectionOutput{
		Entities: entities,
	}, nil
}

// Returns an instance of NamedEntityRepoInterface
func NewNamedEntityRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.NamedEntityRepoInterface {
	metrics := newMetrics(scope)

	return &NamedEntityRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
