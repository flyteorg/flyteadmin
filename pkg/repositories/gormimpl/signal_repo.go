package gormimpl

import (
	"context"
	"errors"

	flyteAdminDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flytestdlib/promutils"

	"gorm.io/gorm"
)

// Implementation of SignalRepoInterface.
type SignalRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (s *SignalRepo) Create(ctx context.Context, input models.Signal) error {
	timer := s.metrics.CreateDuration.Start()
	tx := s.db.Omit("id").Create(&input)
	timer.Stop()
	if tx.Error != nil {
		return s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (s *SignalRepo) Get(ctx context.Context, input interfaces.GetSignalInput) (models.Signal, error) {
	var signal models.Signal
	timer := s.metrics.GetDuration.Start()
	tx := s.db.Where(&models.Signal{
		SignalKey: models.SignalKey{
			ExecutionKey: models.ExecutionKey{
				Project: input.SignalID.ExecutionId.Project,
				Domain:  input.SignalID.ExecutionId.Domain,
				Name:    input.SignalID.ExecutionId.Name,
			},
			SignalID: input.SignalID.SignalId,
		},
	}).Take(&signal)
	timer.Stop()
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.Signal{}, flyteAdminDbErrors.GetMissingEntityError("SIGNAL", &input.SignalID)
	}
	if tx.Error != nil {
		return models.Signal{}, s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return signal, nil
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
