package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
)

// Transforms a WorkflowCreateRequest to a workflow model
func CreateWorkflowModel(id *core.Identifier, remoteClosureIdentifier string,
	typedInterface *core.TypedInterface, digest []byte) (models.Workflow, error) {
	var serializedTypedInterface []byte
	if typedInterface != nil {
		var err error
		serializedTypedInterface, err = proto.Marshal(typedInterface)
		if err != nil {
			return models.Workflow{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize workflow spec")
		}
	}
	return models.Workflow{
		WorkflowKey: models.WorkflowKey{
			Project: id.Project,
			Domain:  id.Domain,
			Name:    id.Name,
			Version: id.Version,
		},
		TypedInterface:          serializedTypedInterface,
		RemoteClosureIdentifier: remoteClosureIdentifier,
		Digest:                  digest,
	}, nil
}

func FromWorkflowModel(workflowModel models.Workflow) (admin.Workflow, error) {
	id := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      workflowModel.WorkflowKey.Project,
		Domain:       workflowModel.WorkflowKey.Domain,
		Name:         workflowModel.WorkflowKey.Name,
		Version:      workflowModel.WorkflowKey.Version,
	}
	createdAt, err := ptypes.TimestampProto(workflowModel.CreatedAt)
	if err != nil {
		return admin.Workflow{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read created at timestamp")
	}

	// Because the spec if offloaded, it is not populated in the model returned here.
	return admin.Workflow{
		Id: &id,
		Closure: &admin.WorkflowClosure{
			CreatedAt: createdAt,
		},
	}, nil
}

func FromWorkflowModels(workflowModels []models.Workflow) ([]*admin.Workflow, error) {
	workflows := make([]*admin.Workflow, len(workflowModels))
	for idx, workflowModel := range workflowModels {
		workflow, err := FromWorkflowModel(workflowModel)
		if err != nil {
			return nil, err
		}
		workflows[idx] = &workflow
	}
	return workflows, nil
}

func FromWorkflowModelsToIdentifiers(workflowModels []models.Workflow) []*admin.NamedEntityIdentifier {
	ids := make([]*admin.NamedEntityIdentifier, len(workflowModels))
	for i, wf := range workflowModels {
		ids[i] = &admin.NamedEntityIdentifier{
			Project: wf.Project,
			Domain:  wf.Domain,
			Name:    wf.Name,
		}
	}

	return ids
}
