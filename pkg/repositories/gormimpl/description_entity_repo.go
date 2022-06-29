package gormimpl

import (
	"context"

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

func (r *DescriptionEntityRepo) Create(ctx context.Context, input models.DescriptionEntity) error {
	timer := r.metrics.CreateDuration.Start()
	tx := r.db.Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *DescriptionEntityRepo) Get(ctx context.Context, input models.DescriptionEntityKey) (models.DescriptionEntity, error) {
	var descriptionEntity models.DescriptionEntity
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.DescriptionEntity{
		DescriptionEntityKey: models.DescriptionEntityKey{
			ResourceType: input.ResourceType,
			Project:      input.Project,
			Domain:       input.Domain,
			Name:         input.Name,
			Version:      input.Version,
		},
	}).Take(&descriptionEntity)
	timer.Stop()
	if tx.Error != nil {
		return models.DescriptionEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return descriptionEntity, nil
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
