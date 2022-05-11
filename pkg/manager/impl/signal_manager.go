package impl

import (
	"context"

	"github.com/flyteorg/flytestdlib/contextutils"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	//"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	//"github.com/prometheus/client_golang/prometheus"
	//"google.golang.org/grpc/codes"
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

	// TODO hamersaw - Assert that a matching signal doesn't already exist before creating the signal.
	/*existingMatchingWorkflow, err := util.GetWorkflowModel(ctx, w.db, *request.Id)
	// Check that no identical or conflicting workflows exist.
	if err == nil {
		// A workflow's structure is uniquely defined by its collection of nodes.
		if bytes.Equal(workflowDigest, existingMatchingWorkflow.Digest) {
			return nil, errors.NewFlyteAdminErrorf(
				codes.AlreadyExists, "identical workflow already exists with id %v", request.Id)
		}
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"workflow with different structure already exists with id %v", request.Id)
	} else if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
		logger.Debugf(ctx, "Failed to get workflow for comparison in CreateWorkflow with ID [%+v] with err %v",
			request.Id, err)
		return nil, err
	}*/

	// Save the signal to the database.
	signalModel, err := transformers.CreateSignalModel(request)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform signal model for request [%+v] with err: %v",
			request, err)
		return nil, err
	}
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
	// TODO hamersaw - impelemnt getSignal from SignalRepo
	/*signal, err := util.GetSignal(ctx, w.db, w.storageClient, *request.Id)
	if err != nil {
		logger.Infof(ctx, "Failed to get workflow with id [%+v] with err %v", request.Id, err)
		return nil, err
	}
	return workflow, nil*/
	return &admin.Signal{}, nil
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
