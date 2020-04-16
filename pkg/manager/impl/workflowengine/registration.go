package workflowengine

import (
	"bytes"
	"context"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/util"
	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

func CreateWorkflow(
	ctx context.Context,
	request admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
	if err := validation.ValidateWorkflow(ctx, request, w.db, w.config.ApplicationConfiguration()); err != nil {
		return nil, err
	}
	ctx = getWorkflowContext(ctx, request.Id)
	finalizedRequest, err := w.setDefaults(request)
	if err != nil {
		logger.Debugf(ctx, "Failed to set defaults for workflow with id [%+v] with err %v", request.Id, err)
		return nil, err
	}
	// Validate that the workflow compiles.
	workflowClosure, err := w.getCompiledWorkflow(ctx, finalizedRequest)
	if err != nil {
		logger.Errorf(ctx, "Failed to compile workflow with err: %v", err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to compile workflow for [%+v] with err %v", request.Id, err)
	}
	err = validation.ValidateCompiledWorkflow(
		*request.Id, workflowClosure, w.config.RegistrationValidationConfiguration())
	if err != nil {
		return nil, err
	}
	workflowDigest, err := util.GetWorkflowDigest(ctx, workflowClosure.CompiledWorkflow)
	if err != nil {
		logger.Errorf(ctx, "failed to compute workflow digest with err %v", err)
		return nil, err
	}

	// Assert that a matching workflow doesn't already exist before uploading the workflow closure.
	existingMatchingWorkflow, err := util.GetWorkflowModel(ctx, w.db, *request.Id)
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
	}

	remoteClosureDataRef, err := w.createDataReference(ctx, request.Spec.Template.Id)
	if err != nil {
		logger.Infof(ctx, "failed to construct data reference for workflow closure with id [%+v] with err %v",
			request.Id, err)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to construct data reference for workflow closure with id [%+v] and err %v", request.Id, err)
	}
	err = w.storageClient.WriteProtobuf(ctx, remoteClosureDataRef, defaultStorageOptions, &workflowClosure)

	if err != nil {
		logger.Infof(ctx,
			"failed to write marshaled workflow with id [%+v] to storage %s with err %v and base container: %s",
			request.Id, remoteClosureDataRef.String(), err, w.storageClient.GetBaseContainerFQN(ctx))
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to write marshaled workflow [%+v] to storage %s with err %v and base container: %s",
			request.Id, remoteClosureDataRef.String(), err, w.storageClient.GetBaseContainerFQN(ctx))
	}
	// Save the workflow & its reference to the offloaded, compiled workflow in the database.
	workflowModel, err := transformers.CreateWorkflowModel(
		finalizedRequest, remoteClosureDataRef.String(), workflowDigest)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform workflow model for request [%+v] and remoteClosureIdentifier [%s] with err: %v",
			finalizedRequest, remoteClosureDataRef.String(), err)
		return nil, err
	}
	if err = w.db.WorkflowRepo().Create(ctx, workflowModel); err != nil {
		logger.Infof(ctx, "Failed to create workflow model [%+v] with err %v", request.Id, err)
		return nil, err
	}
	w.metrics.TypedInterfaceSizeBytes.Observe(float64(len(workflowModel.TypedInterface)))
	return &admin.WorkflowCreateResponse{}, nil
}