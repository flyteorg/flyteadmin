package resources

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	repo_interface "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
)

type ResourceManager struct {
	db     repo_interface.Repository
	config runtimeInterfaces.Configuration
}

func (m *ResourceManager) GetResource(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
	resource, err := m.db.ResourceRepo().Get(ctx, repo_interface.ResourceID{
		ResourceType: request.ResourceType.String(),
		Project:      request.Project,
		Domain:       request.Domain,
		Workflow:     request.Workflow,
		LaunchPlan:   request.LaunchPlan,
	})
	if err != nil {
		return nil, err
	}

	var attributes admin.MatchingAttributes
	err = proto.Unmarshal(resource.Attributes, &attributes)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "Failed to decode resource attribute with err: %v", err)
	}
	return &interfaces.ResourceResponse{
		ResourceType: resource.ResourceType,
		Project:      resource.Project,
		Domain:       resource.Domain,
		Workflow:     resource.Workflow,
		LaunchPlan:   resource.LaunchPlan,
		Attributes:   &attributes,
	}, nil
}

func (m *ResourceManager) GetResourcesList(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponseList, error) {
	resources, err := m.db.ResourceRepo().GetRows(ctx, repo_interface.ResourceID{
		ResourceType: request.ResourceType.String(),
		Project:      request.Project,
		Domain:       request.Domain,
		Workflow:     request.Workflow,
		LaunchPlan:   request.LaunchPlan,
	})
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "Retrieved %d rows listing resource type %s", len(resources), request.ResourceType.String())

	var attributes = make([]*admin.MatchingAttributes, 0, len(resources))
	for _, resource := range resources {
		var attr admin.MatchingAttributes
		err = proto.Unmarshal(resource.Attributes, &attr)
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to decode resource attribute with err: %v", err)
		}
		attributes = append(attributes, &attr)
	}

	return &interfaces.ResourceResponseList{
		ResourceType:  request.ResourceType.String(),
		Project:       request.Project,
		Domain:        request.Domain,
		Workflow:      request.Workflow,
		LaunchPlan:    request.LaunchPlan,
		AttributeList: attributes,
	}, nil
}

func (m *ResourceManager) createOrMergeUpdateWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesUpdateRequest, model models.Resource,
	resourceType admin.MatchableResource) (*admin.WorkflowAttributesUpdateResponse, error) {
	resourceID := repo_interface.ResourceID{
		Project:      model.Project,
		Domain:       model.Domain,
		Workflow:     model.Workflow,
		LaunchPlan:   model.LaunchPlan,
		ResourceType: model.ResourceType,
	}
	existing, err := m.db.ResourceRepo().GetRaw(ctx, resourceID)
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound {
			// Proceed with the default CreateOrUpdate call since there's no existing model to update.
			err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
			if err != nil {
				return nil, err
			}
			return &admin.WorkflowAttributesUpdateResponse{}, nil
		}
		return nil, err
	}
	updatedModel, err := transformers.MergeUpdateWorkflowAttributes(
		ctx, existing, resourceType, &resourceID, request.Attributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, updatedModel)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) UpdateWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateWorkflowAttributesUpdateRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}

	model, err := transformers.WorkflowAttributesToResourceModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	if request.Attributes.GetMatchingAttributes().GetPluginOverrides() != nil {
		return m.createOrMergeUpdateWorkflowAttributes(ctx, request, model, admin.MatchableResource_PLUGIN_OVERRIDE)
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) GetWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {

	// if the request is a task resource request, then call that logic designed to merge task resources from
	// different levels along with base config
	if request.ResourceType == admin.MatchableResource_TASK_RESOURCE {
		r := repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, Workflow: request.Workflow, ResourceType: request.ResourceType.String()}
		matchingAttributes, err := m.HandleGetTaskResourceRequest(ctx, r)
		if err != nil {
			return nil, err
		}
		return &admin.WorkflowAttributesGetResponse{
			Attributes: &admin.WorkflowAttributes{
				Project:            request.Project,
				Domain:             request.Domain,
				Workflow:           request.Workflow,
				MatchingAttributes: matchingAttributes,
			},
		}, nil
	}

	if err := validation.ValidateWorkflowAttributesGetRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}
	workflowAttributesModel, err := m.db.ResourceRepo().Get(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, Workflow: request.Workflow, ResourceType: request.ResourceType.String()})
	if err != nil {
		return nil, err
	}
	workflowAttributes, err := transformers.FromResourceModelToWorkflowAttributes(workflowAttributesModel)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesGetResponse{
		Attributes: &workflowAttributes,
	}, nil
}

func (m *ResourceManager) DeleteWorkflowAttributes(ctx context.Context,
	request admin.WorkflowAttributesDeleteRequest) (*admin.WorkflowAttributesDeleteResponse, error) {
	if err := validation.ValidateWorkflowAttributesDeleteRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}
	if err := m.db.ResourceRepo().Delete(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, Workflow: request.Workflow, ResourceType: request.ResourceType.String()}); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted workflow attributes for: %s-%s-%s (%s)", request.Project,
		request.Domain, request.Workflow, request.ResourceType.String())
	return &admin.WorkflowAttributesDeleteResponse{}, nil
}

func (m *ResourceManager) UpdateProjectAttributes(ctx context.Context, request admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {

	var resource admin.MatchableResource
	var err error

	if resource, err = validation.ValidateProjectAttributesUpdateRequest(ctx, m.db, request); err != nil {
		return nil, err
	}
	model, err := transformers.ProjectAttributesToResourceModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}

	if request.Attributes.GetMatchingAttributes().GetPluginOverrides() != nil {
		return m.createOrMergeUpdateProjectAttributes(ctx, request, model, admin.MatchableResource_PLUGIN_OVERRIDE)
	}

	err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) GetProjectAttributesBase(ctx context.Context, request admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {

	if err := validation.ValidateProjectExists(ctx, m.db, request.Project); err != nil {
		return nil, err
	}

	projectAttributesModel, err := m.db.ResourceRepo().GetProjectLevel(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: "", ResourceType: request.ResourceType.String()})
	if err != nil {
		return nil, err
	}

	ma, err := transformers.FromResourceModelToMatchableAttributes(projectAttributesModel)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectAttributesGetResponse{
		Attributes: &admin.ProjectAttributes{
			Project:            request.Project,
			MatchingAttributes: ma.Attributes,
		},
	}, nil
}

// HandleGetTaskResourceRequest needs to merge results from multiple layers in the db, along with configuration value
func (m *ResourceManager) HandleGetTaskResourceRequest(ctx context.Context, request repo_interface.ResourceID) (*admin.MatchingAttributes, error) {
	if err := validation.ValidateProjectExists(ctx, m.db, request.Project); err != nil {
		return nil, err
	}

	var attrs []admin.TaskResourceAttributes
	attrs = []admin.TaskResourceAttributes{}

	rrList, err := m.GetResourcesList(ctx, interfaces.ResourceRequest{
		Project:      request.Project,
		Domain:       request.Domain,
		Workflow:     request.Workflow,
		LaunchPlan:   "",
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})

	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound {
			logger.Debug(ctx, "HandleGetTaskResourceRequest did not find any task resources, falling back")
		} else {
			return nil, err
		}
	} else {
		logger.Debugf(ctx, "HandleGetTaskResourceRequest returned [%d] task resources, combining with config", len(rrList.AttributeList))
		for _, rr := range rrList.AttributeList {
			if rr.GetTaskResourceAttributes() != nil {
				attrs = append(attrs, *rr.GetTaskResourceAttributes())
			}
		}
	}

	attrs = append(attrs, m.config.TaskResourceConfiguration().GetAsAttribute())
	responseAttributes := util.MergeDownTaskResources(attrs...)

	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{
			TaskResourceAttributes: responseAttributes,
		},
	}, nil
}

// GetProjectAttributes combines the call to the database to get the Project level settings with
// Admin server level configuration.
// Note this merge is only done for WorkflowExecutionConfig
// This merge should be done for the following matchable-resource types:
// TASK_RESOURCE, WORKFLOW_EXECUTION_CONFIG
// This code should be removed pending implementation of a complete settings implementation.
func (m *ResourceManager) GetProjectAttributes(ctx context.Context, request admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {

	// if the request is a task resource request, then call that logic designed to merge task resources from
	// different levels along with base config
	if request.ResourceType == admin.MatchableResource_TASK_RESOURCE {
		r := repo_interface.ResourceID{Project: request.Project, Domain: "", ResourceType: request.ResourceType.String()}
		matchingAttributes, err := m.HandleGetTaskResourceRequest(ctx, r)
		if err != nil {
			return nil, err
		}
		return &admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Project:            request.Project,
				MatchingAttributes: matchingAttributes,
			},
		}, nil
	}

	getResponse, err := m.GetProjectAttributesBase(ctx, request)

	// Return as missing if missing and not one of the two matchable resources that are merged with system level config
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound && (request.ResourceType == admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG || request.ResourceType == admin.MatchableResource_TASK_RESOURCE) {
			logger.Debugf(ctx, "Attributes not found, but look for system fallback %s", request.ResourceType.String())
		} else {
			return nil, err
		}
	}

	// Merge with system level config if appropriate
	if request.ResourceType == admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG {
		var responseAttributes *admin.WorkflowExecutionConfig
		configLevelDefaults := m.config.ApplicationConfiguration().GetAsWorkflowExecutionAttribute()
		if getResponse == nil || getResponse.Attributes == nil || getResponse.Attributes.GetMatchingAttributes() == nil || getResponse.Attributes.GetMatchingAttributes().GetWorkflowExecutionConfig() == nil {
			responseAttributes = &configLevelDefaults
		} else {
			logger.Debugf(ctx, "Merging workflow config %s with defaults %s", responseAttributes, configLevelDefaults)
			responseAttributes = getResponse.Attributes.GetMatchingAttributes().GetWorkflowExecutionConfig()
			tmp := util.MergeIntoExecConfig(*responseAttributes, &configLevelDefaults)
			responseAttributes = &tmp
		}
		return &admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Project: request.Project,
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: responseAttributes,
					},
				},
			},
		}, nil
	} else if request.ResourceType == admin.MatchableResource_TASK_RESOURCE {
		// todo: delete this, handled above
		var responseAttributes *admin.TaskResourceAttributes
		configLevelDefaults := m.config.TaskResourceConfiguration().GetAsAttribute()
		if getResponse == nil || getResponse.Attributes == nil || getResponse.Attributes.GetMatchingAttributes() == nil || getResponse.Attributes.GetMatchingAttributes().GetTaskResourceAttributes() == nil {
			responseAttributes = &configLevelDefaults
		} else {
			logger.Debugf(ctx, "Merging taskresources %v with system config %v", responseAttributes, configLevelDefaults)
			responseAttributes = getResponse.Attributes.GetMatchingAttributes().GetTaskResourceAttributes()
			responseAttributes = util.MergeDownTaskResources(*responseAttributes, configLevelDefaults)
		}
		return &admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Project: request.Project,
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_TaskResourceAttributes{
						TaskResourceAttributes: responseAttributes,
					},
				},
			},
		}, nil
	}

	return getResponse, nil
}

func (m *ResourceManager) DeleteProjectAttributes(ctx context.Context, request admin.ProjectAttributesDeleteRequest) (
	*admin.ProjectAttributesDeleteResponse, error) {

	if err := validation.ValidateProjectForUpdate(ctx, m.db, request.Project); err != nil {
		return nil, err
	}
	if err := m.db.ResourceRepo().Delete(
		ctx, repo_interface.ResourceID{Project: request.Project, ResourceType: request.ResourceType.String()}); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted project attributes for: %s-%s (%s)", request.Project, request.ResourceType.String())
	return &admin.ProjectAttributesDeleteResponse{}, nil
}

func (m *ResourceManager) createOrMergeUpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest, model models.Resource,
	resourceType admin.MatchableResource) (*admin.ProjectDomainAttributesUpdateResponse, error) {
	resourceID := repo_interface.ResourceID{
		Project:      model.Project,
		Domain:       model.Domain,
		Workflow:     model.Workflow,
		LaunchPlan:   model.LaunchPlan,
		ResourceType: model.ResourceType,
	}
	existing, err := m.db.ResourceRepo().GetRaw(ctx, resourceID)
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound {
			// Proceed with the default CreateOrUpdate call since there's no existing model to update.
			err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
			if err != nil {
				return nil, err
			}
			return &admin.ProjectDomainAttributesUpdateResponse{}, nil
		}
		return nil, err
	}
	// TODO: does this belong here? feels like the error should be better handled and not returned
	updatedModel, err := transformers.MergeUpdatePluginAttributes(
		ctx, existing, resourceType, &resourceID, request.Attributes.MatchingAttributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, updatedModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) createOrMergeUpdateProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesUpdateRequest, model models.Resource,
	resourceType admin.MatchableResource) (*admin.ProjectAttributesUpdateResponse, error) {

	resourceID := repo_interface.ResourceID{
		Project:      model.Project,
		Domain:       model.Domain,
		Workflow:     model.Workflow,
		LaunchPlan:   model.LaunchPlan,
		ResourceType: model.ResourceType,
	}
	existing, err := m.db.ResourceRepo().GetRaw(ctx, resourceID)
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound {
			// Proceed with the default CreateOrUpdate call since there's no existing model to update.
			err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
			if err != nil {
				return nil, err
			}
			return &admin.ProjectAttributesUpdateResponse{}, nil
		}
		return nil, err
	}
	updatedModel, err := transformers.MergeUpdatePluginAttributes(
		ctx, existing, resourceType, &resourceID, request.Attributes.MatchingAttributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, updatedModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateProjectDomainAttributesUpdateRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Attributes.Project, request.Attributes.Domain)

	model, err := transformers.ProjectDomainAttributesToResourceModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	if request.Attributes.GetMatchingAttributes().GetPluginOverrides() != nil {
		return m.createOrMergeUpdateProjectDomainAttributes(ctx, request, model, admin.MatchableResource_PLUGIN_OVERRIDE)
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) GetProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {

	// if the request is a task resource request, then call that logic designed to merge task resources from
	// different levels along with base config
	if request.ResourceType == admin.MatchableResource_TASK_RESOURCE {
		r := repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, ResourceType: request.ResourceType.String()}
		matchingAttributes, err := m.HandleGetTaskResourceRequest(ctx, r)
		if err != nil {
			return nil, err
		}
		return &admin.ProjectDomainAttributesGetResponse{
			Attributes: &admin.ProjectDomainAttributes{
				Project:            request.Project,
				Domain:             request.Domain,
				MatchingAttributes: matchingAttributes,
			},
		}, nil
	}

	if err := validation.ValidateProjectDomainAttributesGetRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}
	projectAttributesModel, err := m.db.ResourceRepo().Get(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, ResourceType: request.ResourceType.String()})
	if err != nil {
		return nil, err
	}
	projectAttributes, err := transformers.FromResourceModelToProjectDomainAttributes(projectAttributesModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesGetResponse{
		Attributes: &projectAttributes,
	}, nil
}

func (m *ResourceManager) DeleteProjectDomainAttributes(ctx context.Context,
	request admin.ProjectDomainAttributesDeleteRequest) (*admin.ProjectDomainAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectDomainAttributesDeleteRequest(ctx, m.db, m.config.ApplicationConfiguration(), request); err != nil {
		return nil, err
	}
	if err := m.db.ResourceRepo().Delete(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, ResourceType: request.ResourceType.String()}); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted project-domain attributes for: %s-%s (%s)", request.Project,
		request.Domain, request.ResourceType.String())
	return &admin.ProjectDomainAttributesDeleteResponse{}, nil
}

func (m *ResourceManager) ListAll(ctx context.Context, request admin.ListMatchableAttributesRequest) (
	*admin.ListMatchableAttributesResponse, error) {
	if err := validation.ValidateListAllMatchableAttributesRequest(request); err != nil {
		return nil, err
	}
	resources, err := m.db.ResourceRepo().ListAll(ctx, request.ResourceType.String())
	if err != nil {
		return nil, err
	}
	if resources == nil {
		// That's fine - there don't necessarily need to exist overrides in the database
		return &admin.ListMatchableAttributesResponse{}, nil
	}
	configurations, err := transformers.FromResourceModelsToMatchableAttributes(resources)
	if err != nil {
		return nil, err
	}
	return &admin.ListMatchableAttributesResponse{
		Configurations: configurations,
	}, nil
}

func NewResourceManager(db repo_interface.Repository, config runtimeInterfaces.Configuration) interfaces.ResourceInterface {
	return &ResourceManager{
		db:     db,
		config: config,
	}
}
