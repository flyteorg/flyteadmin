package adminservice

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/audit"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) UpdateWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesUpdateResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.update.Time(func() {
		response, err = m.ResourceManager.UpdateWorkflowAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"UpdateWorkflowAttributes",
		map[string]string{
			audit.Project: request.Attributes.Project,
			audit.Domain:  request.Attributes.Domain,
			audit.Name:    request.Attributes.Workflow,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) GetWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesGetResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ResourceManager.GetWorkflowAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"GetWorkflowAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
			audit.Name:    request.Workflow,
		},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteWorkflowAttributes(ctx context.Context, request *admin.WorkflowAttributesDeleteRequest) (
	*admin.WorkflowAttributesDeleteResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.WorkflowAttributesDeleteResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.ResourceManager.DeleteWorkflowAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"DeleteWorkflowAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
			audit.Name:    request.Workflow,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.delete)
	}

	return response, nil
}

func (m *AdminService) UpdateProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesUpdateResponse
	var err error
	m.Metrics.projectDomainAttributesEndpointMetrics.update.Time(func() {
		response, err = m.ResourceManager.UpdateProjectDomainAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"UpdateProjectDomainAttributes",
		map[string]string{
			audit.Project: request.Attributes.Project,
			audit.Domain:  request.Attributes.Domain,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectDomainAttributesEndpointMetrics.update)
	}

	return response, nil
}

func (m *AdminService) GetProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesGetResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.get.Time(func() {
		response, err = m.ResourceManager.GetProjectDomainAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"GetProjectDomainAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
		},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.get)
	}

	return response, nil
}

func (m *AdminService) DeleteProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesDeleteRequest) (
	*admin.ProjectDomainAttributesDeleteResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesDeleteResponse
	var err error
	m.Metrics.workflowAttributesEndpointMetrics.delete.Time(func() {
		response, err = m.ResourceManager.DeleteProjectDomainAttributes(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"DeleteProjectDomainAttributes",
		map[string]string{
			audit.Project: request.Project,
			audit.Domain:  request.Domain,
		},
		audit.ReadWrite,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.workflowAttributesEndpointMetrics.delete)
	}

	return response, nil
}

func (m *AdminService) ListMatchableAttributes(ctx context.Context, request *admin.ListMatchableAttributesRequest) (
	*admin.ListMatchableAttributesResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ListMatchableAttributesResponse
	var err error
	m.Metrics.matchableAttributesEndpointMetrics.list.Time(func() {
		response, err = m.ResourceManager.ListAll(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"ListMatchableAttributes",
		map[string]string{
			audit.ResourceType: request.ResourceType.String(),
		},
		audit.ReadOnly,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.matchableAttributesEndpointMetrics.list)
	}

	return response, nil
}
