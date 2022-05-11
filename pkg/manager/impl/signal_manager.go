package impl

import (
	"bytes"
	"context"

	"github.com/flyteorg/flytestdlib/contextutils"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	//"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
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
	return contextutils.WithWorkflowID(ctx, identifier.ExecutionId.Name)
	// TODO hamersaw - add identifier.SignalId
}

func (s *SignalManager) CreateSignal(
	ctx context.Context,
	request admin.SignalCreateRequest) (*admin.SignalCreateResponse, error) {
	// TODO hamersaw - add validation to check the signal type
	/*if err := validation.ValidateWorkflow(ctx, request, w.db, w.config.ApplicationConfiguration()); err != nil {
		return nil, err
	}*/

	ctx = getSignalContext(ctx, request.Id)
	signalModel, err := transformers.CreateSignalModel(request)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform signal model for request [%+v] with err: %v",
			request, err)
		return nil, err
	}

	// Assert that a matching signal doesn't already exist before creating the signal.
	existingSignalModel, err := s.db.SignalRepo().Get(ctx, repoInterfaces.GetSignalInput{*request.Id});
	if err == nil {
		// A signals structure is uniquely defined by its value.
		if bytes.Equal(signalModel.Value, existingSignalModel.Value) {
			return nil, errors.NewFlyteAdminErrorf(
				codes.AlreadyExists, "identical signal already exists with id %v", request.Id)
		}
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"signal with different value already exists with id %v", request.Id)
	} else if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
		logger.Debugf(ctx, "Failed to get signal for comparison in CreateSignal with ID [%+v] with err %v",
			request.Id, err)
		return nil, err
	}

	// Save the signal to the database.
	if err = s.db.SignalRepo().Create(ctx, signalModel); err != nil {
		logger.Infof(ctx, "Failed to create signal model [%+v] with err %v", request.Id, err)
		return nil, err
	}
	return &admin.SignalCreateResponse{}, nil
}

func (s *SignalManager) GetSignal(ctx context.Context, request admin.SignalGetRequest) (*admin.Signal, error) {
	// TODO hamersaw - validate signal
	/*if err := validation.ValidateIdentifier(request.Id, common.Workflow); err != nil {
		logger.Debugf(ctx, "invalid identifier [%+v]: %v", request.Id, err)
		return nil, err
	}*/
	ctx = getSignalContext(ctx, request.Id)
	signalModel, err := s.db.SignalRepo().Get(ctx, repoInterfaces.GetSignalInput{*request.Id});
	if err != nil {
		// TODO hamersaw - do we want to log this? are we really failing or does it just not exist yet?
		logger.Infof(ctx, "Failed to get signal model [%+v] with err %v", request.Id, err)
		return nil, err
	}
	signal, err := transformers.FromSignalModel(signalModel)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform signal model [%+v] with err: %v", signalModel, err)
		return nil, err
	}
	return &signal, nil
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
