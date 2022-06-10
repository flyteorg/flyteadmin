package impl

import (
	"context"

	"github.com/flyteorg/flytestdlib/contextutils"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	//"github.com/prometheus/client_golang/prometheus"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
)

type signalMetrics struct {
	Scope promutils.Scope
	// TODO hamersaw - add some signal metrics
}

type SignalManager struct {
	db      repoInterfaces.Repository
	metrics signalMetrics
}

func getSignalContext(ctx context.Context, identifier *core.SignalIdentifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.ExecutionId.Project, identifier.ExecutionId.Domain)
	ctx = contextutils.WithWorkflowID(ctx, identifier.ExecutionId.Name)
	return context.WithValue(ctx, "signal_id", identifier.SignalId)
}

func (s *SignalManager) GetOrCreateSignal(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
	if err := validation.ValidateSignalGetOrCreateRequest(ctx, request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}

	ctx = getSignalContext(ctx, request.Id)
	signalModel, err := transformers.CreateSignalModel(request.Id, request.Type, nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal with id [%+v] and type [+%v] with err: %v", request.Id, request.Type, err)
		return nil, err
	}

	err = s.db.SignalRepo().GetOrCreate(ctx, &signalModel);
	if err != nil {
		return nil, err
	}

	signal, err := transformers.FromSignalModel(signalModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal model [%+v] with err: %v", signalModel, err)
		return nil, err
	}

	return &signal, nil
}

func (s *SignalManager) ListSignals(ctx context.Context, request admin.SignalListRequest) ([]*admin.Signal, error) {
	// TODO hamersaw - validate signal
	/*if err := validation.ValidateIdentifier(request.Id, common.Workflow); err != nil {
		logger.Debugf(ctx, "invalid identifier [%+v]: %v", request.Id, err)
		return nil, err
	}*/

	/*ctx = getSignalContext(ctx, request.Id)
	signalModel, err := transformers.CreateSignalModel(*request.Id, request.Type, nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal with id [%+v] and type [+%v] with err: %v", request.Id, request.Type, err)
		return nil, err
	}

	err = s.db.SignalRepo().GetOrCreate(ctx, &signalModel);
	if err != nil {
		return nil, err
	}

	signal, err := transformers.FromSignalModel(signalModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal model [%+v] with err: %v", signalModel, err)
		return nil, err
	}

	return &signal, nil*/
	return nil, nil
}

func (s *SignalManager) SetSignal(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
	if err := validation.ValidateSignalSetRequest(ctx, s.db, request); err != nil {
		return nil, err
	}

	signalModel, err := transformers.CreateSignalModel(request.Id, nil, request.Value)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal with id [%+v] and value [+%v] with err: %v", request.Id, request.Value, err)
		return nil, err
	}

	err = s.db.SignalRepo().Update(ctx, signalModel.SignalKey, signalModel.Value);
	if err != nil {
		return nil, err
	}

	return &admin.SignalSetResponse{}, nil
}

func NewSignalManager(
	db repoInterfaces.Repository,
	scope promutils.Scope) interfaces.SignalInterface {
	metrics := signalMetrics{
		Scope: scope,
	}

	return &SignalManager{
		db:      db,
		metrics: metrics,
	}
}
