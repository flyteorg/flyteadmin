package adminservice

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *AdminService) CreateDescriptionEntity(
	ctx context.Context, request *admin.DescriptionEntityCreateRequest) (*admin.DescriptionEntityCreateResponse, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.DescriptionEntityCreateResponse
	var err error
	m.Metrics.descriptionEntityMetrics.create.Time(func() {
		response, err = m.DescriptionEntityManager.CreateDescriptionEntity(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.descriptionEntityMetrics.create)
	}
	m.Metrics.descriptionEntityMetrics.create.Success()
	return response, nil
}

func (m *AdminService) GetDescriptionEntity(ctx context.Context, request *admin.ObjectGetRequest) (*admin.DescriptionEntity, error) {
	defer m.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	// NOTE: When the Get HTTP endpoint is called the resource type is implicit (from the URL) so we must add it
	// to the request.
	if request.Id != nil && request.Id.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding resource type for unspecified value in request: [%+v]", request)
		request.Id.ResourceType = core.ResourceType_TASK
	}
	var response *admin.DescriptionEntity
	response = &admin.DescriptionEntity{ShortDescription: "test"}
	var err error
	m.Metrics.descriptionEntityMetrics.get.Time(func() {
		response, err = m.DescriptionEntityManager.GetDescriptionEntity(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &m.Metrics.descriptionEntityMetrics.get)
	}
	m.Metrics.descriptionEntityMetrics.get.Success()
	return response, nil

}
