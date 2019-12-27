package impl

import (
	"context"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"

	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type ProjectDomainAttributesManager struct {
	db     repositories.RepositoryInterface
	config runtimeInterfaces.Configuration
}

func (m *ProjectDomainAttributesManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateProjectDomainAttributesUpdateRequest(request); err != nil {
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Attributes.Project, request.Attributes.Domain)

	model, err := transformers.ToProjectDomainAttributesModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	err = m.db.ProjectDomainAttributesRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func NewProjectDomainAttributesManager(
	db repositories.RepositoryInterface, config runtimeInterfaces.Configuration) interfaces.ProjectDomainAttributesInterface {
	return &ProjectDomainAttributesManager{
		db:     db,
		config: config,
	}
}
