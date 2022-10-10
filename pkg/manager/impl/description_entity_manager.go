package impl

import (
	"bytes"
	"context"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"google.golang.org/grpc/codes"
)

type DescriptionEntityMetrics struct {
	Scope promutils.Scope
}

type DescriptionEntityManager struct {
	db      repoInterfaces.Repository
	config  runtimeInterfaces.Configuration
	metrics DescriptionEntityMetrics
}

func createDescriptionEntity(ctx context.Context, db repoInterfaces.Repository, descriptionEntity *admin.DescriptionEntity, id core.Identifier) error {
	descriptionDigest, err := util.GetDescriptionEntityDigest(ctx, descriptionEntity)
	if err != nil {
		logger.Errorf(ctx, "failed to compute description entity digest for [%+v] with err: %v", id, err)
		return err
	}

	existingDescriptionEntityModel, err := util.GetDescriptionEntityModel(ctx, db, id)
	if err == nil {
		if bytes.Equal(existingDescriptionEntityModel.Digest, descriptionDigest) {
			return errors.NewFlyteAdminErrorf(codes.AlreadyExists,
				"identical description entity already exists with id %v", id)
		}

		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"description entity with different structure already exists with id %v", id)
	}

	descriptionModel, err := transformers.CreateDescriptionEntityModel(descriptionEntity, id, descriptionDigest)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform description model [%+v] with err: %v", descriptionEntity, err)
		return err
	}

	var descriptionID uint
	if descriptionID, err = db.DescriptionEntityRepo().Create(ctx, descriptionModel); err != nil {
		logger.Errorf(ctx, "Failed to create description model with id [%+v] with err %v", id, err)
		return err
	}

	if id.ResourceType == core.ResourceType_TASK {
		err = db.TaskRepo().UpdateDescriptionID(models.Task{
			TaskKey: models.TaskKey{
				Project: descriptionModel.Project,
				Domain:  descriptionModel.Domain,
				Name:    descriptionModel.Name,
				Version: descriptionModel.Version,
			},
			DescriptionID: descriptionID,
		})
		if err != nil {
			logger.Errorf(ctx, "Failed to update descriptionID in tasks table: %v", err)
			return err
		}
	} else if id.ResourceType == core.ResourceType_WORKFLOW {
		err = db.WorkflowRepo().UpdateDescriptionID(models.Workflow{
			WorkflowKey: models.WorkflowKey{
				Project: descriptionModel.Project,
				Domain:  descriptionModel.Domain,
				Name:    descriptionModel.Name,
				Version: descriptionModel.Version,
			},
			DescriptionID: descriptionID,
		})
		if err != nil {
			logger.Errorf(ctx, "Failed to update descriptionID in workflows table: %v", err)
			return err
		}
	}

	return nil
}

func (d *DescriptionEntityManager) GetDescriptionEntity(ctx context.Context, request admin.ObjectGetRequest) (
	*admin.DescriptionEntity, error) {
	if err := validation.ValidateDescriptionEntityGetRequest(request); err != nil {
		logger.Errorf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)
	return util.GetDescriptionEntity(ctx, d.db, *request.Id)
}

func (d *DescriptionEntityManager) ListDescriptionEntity(ctx context.Context, request admin.DescriptionEntityListRequest) (*admin.DescriptionEntityList, error) {
	// Check required fields
	if err := validation.ValidateDescriptionEntityListRequest(request); err != nil {
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.DescriptionEntityId.Project, request.DescriptionEntityId.Domain)
	ctx = contextutils.WithWorkflowID(ctx, request.DescriptionEntityId.Name)
	filters, err := util.GetDbFilters(util.FilterSpec{
		Project:        request.DescriptionEntityId.Project,
		Domain:         request.DescriptionEntityId.Domain,
		Name:           request.DescriptionEntityId.Name,
		RequestFilters: request.Filters,
	}, common.ResourceTypeToEntity[request.DescriptionEntityId.ResourceType])
	if err != nil {
		logger.Error(ctx, "failed to get database filter")
		return nil, err
	}
	var sortParameter common.SortParameter
	if request.SortBy != nil {
		sortParameter, err = common.NewSortParameter(*request.SortBy)
		if err != nil {
			return nil, err
		}
	}
	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListWorkflows", request.Token)
	}
	listDescriptionEntitiesInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}
	output, err := d.db.DescriptionEntityRepo().List(ctx, listDescriptionEntitiesInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list workflows with [%+v] with err %v", request.DescriptionEntityId, err)
		return nil, err
	}
	descriptionEntityList, err := transformers.FromDescriptionEntityModels(output.Entities)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform workflow models [%+v] with err: %v", output.Entities, err)
		return nil, err
	}
	var token string
	if len(output.Entities) == int(request.Limit) {
		token = strconv.Itoa(offset + len(output.Entities))
	}
	return &admin.DescriptionEntityList{
		DescriptionEntities: descriptionEntityList,
		Token:               token,
	}, nil
}

func NewDescriptionEntityManager(
	db repoInterfaces.Repository,
	config runtimeInterfaces.Configuration,
	scope promutils.Scope) interfaces.DescriptionEntityInterface {

	metrics := DescriptionEntityMetrics{
		Scope: scope,
	}
	return &DescriptionEntityManager{
		db:      db,
		config:  config,
		metrics: metrics,
	}
}
