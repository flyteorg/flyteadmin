package gormimpl

import (
	"context"
	"fmt"

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
	var metadata models.NamedEntityMetadata
	timer := r.metrics.GetDuration.Start()

	// TODO (randy): Verify that a record with the named entity key exists.
	// Return empty if it doesn't
	tx := r.db.Where(&models.NamedEntityMetadata{
		NamedEntityMetadataKey: models.NamedEntityMetadataKey{
			ResourceType: input.ResourceType,
			Project:      input.Project,
			Domain:       input.Domain,
			Name:         input.Name,
		},
	}).First(&metadata)
	timer.Stop()

	// If a record is not found, we will return empty metadata. Otherwise
	// return the error.
	if tx.Error != nil && !tx.RecordNotFound() {
		return models.NamedEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return models.NamedEntity{
		NamedEntityKey: models.NamedEntityKey{
			ResourceType: input.ResourceType,
			Project:      input.Project,
			Domain:       input.Domain,
			Name:         input.Name,
		},
		NamedEntityMetadataFields: metadata.NamedEntityMetadataFields,
	}, nil
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

	groupBy := fmt.Sprintf("%s.%s, %s.%s, %s.%s, %s.%s", tableName, Project, tableName, Domain, tableName, Name, namedEntityMetadataTableName, Description)

	// Scan the results into a list of named entities
	var entities []models.NamedEntity
	timer := r.metrics.ListDuration.Start()
	tx.Select([]string{
		fmt.Sprintf("%s.%s", tableName, Project),
		fmt.Sprintf("%s.%s", tableName, Domain),
		fmt.Sprintf("%s.%s", tableName, Name),
		fmt.Sprintf("%s.%s", namedEntityMetadataTableName, Description),
	}).Group(groupBy).Scan(&entities)
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
