package gormimpl

import (
	"context"

	flyteAdminDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flytestdlib/promutils"

	"gorm.io/gorm"
)

// SignalRepo is an implementation of SignalRepoInterface.
type SignalRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

// GetOrCreate returns a signal if it already exists, if not it creates a new one given the input
func (s *SignalRepo) GetOrCreate(ctx context.Context, input *models.Signal) error {
	timer := s.metrics.CreateDuration.Start()
	tx := s.db.FirstOrCreate(&input, input)
	timer.Stop()
	if tx.Error != nil {
		return s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

// List fetches all signals that match the provided input
func (s *SignalRepo) List(ctx context.Context, input models.Signal) ([]*models.Signal, error) {
	var signals []*models.Signal
	timer := s.metrics.ListDuration.Start()
	tx := s.db.Where(&input).Find(&signals)
	timer.Stop()
	if tx.Error != nil {
		return nil, s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return signals, nil
}

// Update sets the value field on the specified signal model
func (s *SignalRepo) Update(ctx context.Context, input models.Signal) error {
	timer := s.metrics.GetDuration.Start()
	updateTx := s.db.Model(&input).Select("value").Updates(input)
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
