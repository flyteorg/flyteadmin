package adminservice

import (
	"context"
	"time"

	"github.com/lyft/flyteadmin/pkg/audit"

	"github.com/lyft/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) UpdateProjectDomainAttributes(ctx context.Context, request *admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	defer m.interceptPanic(ctx, request)
	requestedAt := time.Now()
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.ProjectDomainAttributesUpdateResponse
	var err error
	m.Metrics.projectEndpointMetrics.register.Time(func() {
		response, err = m.ProjectDomainManager.UpdateProjectDomain(ctx, *request)
	})
	audit.NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"CreateNodeEvent",
		map[string]string{
			audit.Project: request.Attributes.Project,
			audit.Domain:  request.Attributes.Domain,
		},
		admin.Request_READ_WRITE,
		requestedAt,
	).WithResponse(time.Now(), err).Log(ctx)
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.projectDomainEndpointMetrics.update)
	}

	return response, nil
}
