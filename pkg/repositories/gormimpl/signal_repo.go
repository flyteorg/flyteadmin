package gormimpl

import (
	"context"
	//"errors"

	flyteAdminDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flytestdlib/promutils"

	"gorm.io/gorm"
	//"gorm.io/gorm/clause"
)

// Implementation of SignalRepoInterface.
type SignalRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (s *SignalRepo) GetOrCreate(ctx context.Context, input *models.Signal) error {
	timer := s.metrics.CreateDuration.Start()
	tx := s.db.FirstOrCreate(&input, input)
	timer.Stop()
	if tx.Error != nil {
		return s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (s *SignalRepo) Update(ctx context.Context, signal models.Signal) error {
	timer := s.metrics.GetDuration.Start()
	updateTx := s.db.Model(&signal).Omit("type").Updates(signal)
	timer.Stop()
	if updateTx.Error != nil {
		return s.errorTransformer.ToFlyteAdminError(updateTx.Error)
	}
	return nil
}

// Returns an instance of SignalRepoInterface
func NewSignalRepo(
	db *gorm.DB, errorTransformer flyteAdminDbErrors.ErrorTransformer, scope promutils.Scope) interfaces.SignalRepoInterface {
	metrics := newMetrics(scope)
	return &SignalRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
