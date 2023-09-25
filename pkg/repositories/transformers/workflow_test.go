package transformers

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/stretchr/testify/assert"
)

const remoteClosureIdentifier = "remote closure id"

var workflowDigest = []byte("workflow digest")

func TestCreateWorkflow(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	workflow, err := CreateWorkflowModel(request, remoteClosureIdentifier, workflowDigest)
	assert.NoError(t, err)
	assert.Equal(t, "project", workflow.Project)
	assert.Equal(t, "domain", workflow.Domain)
	assert.Equal(t, "name", workflow.Name)
	assert.Equal(t, "version", workflow.Version)
	expectedTypedInterface := testutils.GetWorkflowRequestInterfaceBytes()
	assert.Equal(t, expectedTypedInterface, workflow.TypedInterface)
	assert.Equal(t, remoteClosureIdentifier, workflow.RemoteClosureIdentifier)
	assert.Equal(t, workflowDigest, workflow.Digest)
}

func TestCreateWorkflowEmptyInterface(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	request.Spec.Template.Interface = nil
	workflow, err := CreateWorkflowModel(request, remoteClosureIdentifier, workflowDigest)
	assert.NoError(t, err)
	assert.Equal(t, "project", workflow.Project)
	assert.Equal(t, "domain", workflow.Domain)
	assert.Equal(t, "name", workflow.Name)
	assert.Equal(t, "version", workflow.Version)
	assert.Empty(t, workflow.TypedInterface)
	assert.Equal(t, remoteClosureIdentifier, workflow.RemoteClosureIdentifier)
	assert.Equal(t, workflowDigest, workflow.Digest)
}

func TestFromWorkflowModel(t *testing.T) {
	createdAt := time.Now()
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	workflowModel := models.Workflow{
		BaseModel: models.BaseModel{
			CreatedAt: createdAt,
		},
		WorkflowKey: models.WorkflowKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
		TypedInterface:          testutils.GetWorkflowRequestInterfaceBytes(),
		RemoteClosureIdentifier: remoteClosureIdentifier,
	}
	workflow, err := FromWorkflowModel(workflowModel)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}, workflow.Id))

	var workflowInterface core.TypedInterface
	err = proto.Unmarshal(workflowModel.TypedInterface, &workflowInterface)
	assert.NoError(t, err)

	assert.True(t, proto.Equal(&admin.WorkflowClosure{
		CreatedAt: createdAtProto,
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Interface: &workflowInterface,
				},
			},
		},
	}, workflow.Closure))
}

func TestFromWorkflowModels(t *testing.T) {
	createdAtA := time.Now()
	createdAtAProto, _ := ptypes.TimestampProto(createdAtA)

	createdAtB := createdAtA.Add(time.Hour)
	createdAtBProto, _ := ptypes.TimestampProto(createdAtB)

	workflowModels := []models.Workflow{
		{
			BaseModel: models.BaseModel{
				CreatedAt: createdAtA,
			},
			WorkflowKey: models.WorkflowKey{
				Project: "project a",
				Domain:  "domain a",
				Name:    "name a",
				Version: "version a",
			},
			TypedInterface:          testutils.GetWorkflowRequestInterfaceBytes(),
			RemoteClosureIdentifier: remoteClosureIdentifier,
		},
		{
			BaseModel: models.BaseModel{
				CreatedAt: createdAtB,
			},
			WorkflowKey: models.WorkflowKey{
				Project: "project b",
				Domain:  "domain b",
				Name:    "name b",
				Version: "version b",
			},
		},
	}

	workflowList, err := FromWorkflowModels(workflowModels)
	assert.NoError(t, err)
	assert.Len(t, workflowList, len(workflowModels))
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project a",
		Domain:       "domain a",
		Name:         "name a",
		Version:      "version a",
	}, workflowList[0].Id))

	var workflowInterface0 core.TypedInterface
	err = proto.Unmarshal(workflowModels[0].TypedInterface, &workflowInterface0)
	assert.NoError(t, err)

	assert.True(t, proto.Equal(&admin.WorkflowClosure{
		CreatedAt: createdAtAProto,
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Interface: &workflowInterface0,
				},
			},
		},
	}, workflowList[0].Closure))

	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project b",
		Domain:       "domain b",
		Name:         "name b",
		Version:      "version b",
	}, workflowList[1].Id))

	// Expected to be nil
	var workflowInterface1 core.TypedInterface
	err = proto.Unmarshal(workflowModels[1].TypedInterface, &workflowInterface1)
	assert.NoError(t, err)

	assert.True(t, proto.Equal(&admin.WorkflowClosure{
		CreatedAt: createdAtBProto,
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Interface: &workflowInterface1,
				},
			},
		},
	}, workflowList[1].Closure))
}
