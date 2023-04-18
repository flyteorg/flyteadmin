package util

import (
	"context"
	"testing"

	managerInterfaces "github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	managerMocks "github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	workflowengineInterfaces "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/api/resource"
)

var workflowIdentifier = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      "project",
	Domain:       "domain",
	Name:         "name",
	Version:      "version",
}

// GetPtr because golang
func GetPtr(quantity resource.Quantity) *resource.Quantity {
	return &quantity
}

func TestGetTaskResources(t *testing.T) {
	taskConfig := runtimeMocks.MockTaskResourceConfiguration{}
	taskConfig.Defaults = runtimeInterfaces.TaskResourceSet{
		CPU:              GetPtr(resource.MustParse("200m")),
		GPU:              GetPtr(resource.MustParse("8")),
		Memory:           GetPtr(resource.MustParse("200Gi")),
		EphemeralStorage: GetPtr(resource.MustParse("500Mi")),
	}
	taskConfig.Limits = runtimeInterfaces.TaskResourceSet{
		CPU:              GetPtr(resource.MustParse("300m")),
		GPU:              GetPtr(resource.MustParse("8")),
		Memory:           GetPtr(resource.MustParse("500Gi")),
		EphemeralStorage: GetPtr(resource.MustParse("501Mi")),
	}

	t.Run("use runtime application values", func(t *testing.T) {
		resourceManager := managerMocks.MockResourceManager{}
		resourceManager.GetResourceFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponse, error) {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				Workflow:     workflowIdentifier.Name,
				ResourceType: admin.MatchableResource_TASK_RESOURCE,
			})
			return &managerInterfaces.ResourceResponse{}, nil
		}

		taskResourceAttrs := GetTaskResources(context.TODO(), &workflowIdentifier, &resourceManager, &taskConfig)

		assert.EqualValues(t, taskResourceAttrs, workflowengineInterfaces.TaskResources{
			Defaults: runtimeInterfaces.TaskResourceSet{
				CPU:              GetPtr(resource.MustParse("200m")),
				GPU:              GetPtr(resource.MustParse("8")),
				Memory:           GetPtr(resource.MustParse("200Gi")),
				EphemeralStorage: GetPtr(resource.MustParse("500Mi")),
			},
			Limits: runtimeInterfaces.TaskResourceSet{
				CPU:              GetPtr(resource.MustParse("300m")),
				GPU:              GetPtr(resource.MustParse("8")),
				Memory:           GetPtr(resource.MustParse("500Gi")),
				EphemeralStorage: GetPtr(resource.MustParse("501Mi")),
			},
		})
	})

	t.Run("use specific overrides", func(t *testing.T) {
		resourceManager := managerMocks.MockResourceManager{}
		resourceManager.GetResourcesListFunc = func(ctx context.Context,
			request managerInterfaces.ResourceRequest) (*managerInterfaces.ResourceResponseList, error) {
			assert.EqualValues(t, request, managerInterfaces.ResourceRequest{
				Project:      workflowIdentifier.Project,
				Domain:       workflowIdentifier.Domain,
				Workflow:     workflowIdentifier.Name,
				ResourceType: admin.MatchableResource_TASK_RESOURCE,
			})
			return &managerInterfaces.ResourceResponseList{
				AttributeList: []*admin.MatchingAttributes{
					{
						Target: &admin.MatchingAttributes_TaskResourceAttributes{
							TaskResourceAttributes: &admin.TaskResourceAttributes{
								Defaults: &admin.TaskResourceSpec{
									Cpu:              "1200m",
									Gpu:              "18",
									Memory:           "1200Gi",
									EphemeralStorage: "1500Mi",
								},
								Limits: &admin.TaskResourceSpec{
									Cpu:              "300m",
									Gpu:              "8",
									Memory:           "500Gi",
									EphemeralStorage: "501Mi",
								},
							},
						},
					},
				},
			}, nil
		}
		taskResourceAttrs := GetTaskResources(context.TODO(), &workflowIdentifier, &resourceManager, &taskConfig)
		assert.EqualValues(t, taskResourceAttrs, workflowengineInterfaces.TaskResources{
			Defaults: runtimeInterfaces.TaskResourceSet{
				CPU:              GetPtr(resource.MustParse("300m")),
				GPU:              GetPtr(resource.MustParse("8")),
				Memory:           GetPtr(resource.MustParse("500Gi")),
				EphemeralStorage: GetPtr(resource.MustParse("501Mi")),
			},
			Limits: runtimeInterfaces.TaskResourceSet{
				CPU:              GetPtr(resource.MustParse("300m")),
				GPU:              GetPtr(resource.MustParse("8")),
				Memory:           GetPtr(resource.MustParse("500Gi")),
				EphemeralStorage: GetPtr(resource.MustParse("501Mi")),
			},
		})
	})
}

func TestFromAdminProtoTaskResourceSpec(t *testing.T) {
	taskResourceSet := fromAdminProtoTaskResourceSpec(context.TODO(), &admin.TaskResourceSpec{
		Cpu:              "1",
		Memory:           "100",
		EphemeralStorage: "300",
		Gpu:              "2",
	})
	assert.EqualValues(t, runtimeInterfaces.TaskResourceSet{
		CPU:              GetPtr(resource.MustParse("1")),
		Memory:           GetPtr(resource.MustParse("100")),
		EphemeralStorage: GetPtr(resource.MustParse("300")),
		GPU:              GetPtr(resource.MustParse("2")),
	}, taskResourceSet)
}

func TestGetTaskResourcesAsSet(t *testing.T) {
	taskResources := getTaskResourcesAsSet(context.TODO(), []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "100",
		},
		{
			Name:  core.Resources_MEMORY,
			Value: "200",
		},
		{
			Name:  core.Resources_EPHEMERAL_STORAGE,
			Value: "300",
		},
		{
			Name:  core.Resources_GPU,
			Value: "400",
		},
	})
	assert.True(t, taskResources.CPU.Equal(resource.MustParse("100")))
	assert.True(t, taskResources.Memory.Equal(resource.MustParse("200")))
	assert.True(t, taskResources.EphemeralStorage.Equal(resource.MustParse("300")))
	assert.True(t, taskResources.GPU.Equal(resource.MustParse("400")))
}

// TODO add test with default limits
