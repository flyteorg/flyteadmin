package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
)

var testLaunchPlanDigest = []byte{0x2, 0x46, 0x44, 0x52, 0x5e, 0x5c, 0x62, 0xe7, 0xe0, 0xb1, 0x19, 0x76, 0xf7, 0xaf, 0x28, 0x2b, 0xc, 0x22, 0x3b, 0xc7, 0x21, 0xba, 0x4d, 0x98, 0x5, 0xd9, 0x17, 0x79, 0xd5, 0xb5, 0x2f, 0xf2}

var taskIdentifier = core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      "project",
	Domain:       "domain",
	Name:         "name",
	Version:      "version",
}

var compiledTask = &core.CompiledTask{
	Template: &core.TaskTemplate{
		Id:   &taskIdentifier,
		Type: "foo type",
		Metadata: &core.TaskMetadata{
			Timeout: &duration.Duration{
				Seconds: 60,
			},
		},
		Custom: &_struct.Struct{},
	},
}

var compiledTaskDigest = []byte{
	0x85, 0xb2, 0x84, 0xaa, 0x87, 0x26, 0xa1, 0x3e, 0xc2, 0x20, 0x53, 0x69, 0x82, 0x81, 0xb1, 0x3f, 0xd8, 0xa8, 0xa5,
	0xa, 0x22, 0x80, 0xb1, 0x8, 0x44, 0x53, 0xf3, 0xca, 0x60, 0x4, 0xf7, 0x6f}

var compiledWorkflowDigest = []byte{0x53, 0xf1, 0xe2, 0x37, 0xa5, 0x65, 0x3a, 0xf4, 0xd, 0x67, 0xda, 0xcd, 0x67, 0x2e, 0x1d, 0x6c, 0xcb, 0xdd, 0x40, 0xb7, 0xf8, 0x9, 0xa1, 0x6d, 0x88, 0x61, 0xed, 0xc2, 0x8c, 0xe9, 0x55, 0x69}

func getLaunchPlan() *admin.LaunchPlan {
	return &admin.LaunchPlan{
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs: &core.ParameterMap{
				Parameters: []*core.ParameterMapEntry{
					{
						Name: "foo",
						Var:  &core.Parameter{},
					},
					{
						Name: "bar",
						Var:  &core.Parameter{},
					},
				},
			},
			ExpectedOutputs: &core.VariableMap{
				Variables: []*core.VariableMapEntry{
					{
						Name: "baz",
						Var:  &core.Variable{},
					},
				},
			},
		},
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				ResourceType: core.ResourceType_WORKFLOW,
				Project:      "project",
				Domain:       "domain",
				Name:         "workflow name",
				Version:      "version",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "* * * * ",
					},
				},
			},
		},
	}
}

func getCompiledWorkflow() (*core.CompiledWorkflowClosure, error) {
	var compiledWorkflow core.CompiledWorkflowClosure
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	workflowJSONFile := filepath.Join(pwd, "testdata", "workflow.json")
	workflowJSON, err := ioutil.ReadFile(workflowJSONFile)
	if err != nil {
		return nil, err
	}
	err = jsonpb.UnmarshalString(string(workflowJSON), &compiledWorkflow)
	if err != nil {
		return nil, err
	}
	return &compiledWorkflow, nil
}

func TestGetLaunchPlanDigest(t *testing.T) {
	launchPlanDigest, err := GetLaunchPlanDigest(context.Background(), getLaunchPlan())
	assert.Equal(t, testLaunchPlanDigest, launchPlanDigest)
	assert.Nil(t, err)
}

func TestGetLaunchPlanDigest_Unequal(t *testing.T) {
	launchPlanWithDifferentInputs := getLaunchPlan()
	launchPlanWithDifferentInputs.Closure.ExpectedInputs.Parameters = append(launchPlanWithDifferentInputs.Closure.ExpectedInputs.Parameters, &core.ParameterMapEntry{
		Name: "unexpected",
		Var:  &core.Parameter{},
	})
	launchPlanDigest, err := GetLaunchPlanDigest(context.Background(), launchPlanWithDifferentInputs)
	assert.NotEqual(t, testLaunchPlanDigest, launchPlanDigest)
	assert.Nil(t, err)

	launchPlanWithDifferentOutputs := getLaunchPlan()
	launchPlanWithDifferentOutputs.Closure.ExpectedOutputs.Variables = append(launchPlanWithDifferentOutputs.Closure.ExpectedOutputs.Variables, &core.VariableMapEntry{
		Name: "unexpected",
		Var:  &core.Variable{},
	})
	launchPlanDigest, err = GetLaunchPlanDigest(context.Background(), launchPlanWithDifferentOutputs)
	assert.NotEqual(t, testLaunchPlanDigest, launchPlanDigest)
	assert.Nil(t, err)
}

func TestGetTaskDigest(t *testing.T) {
	taskDigest, err := GetTaskDigest(context.Background(), compiledTask)
	assert.Equal(t, compiledTaskDigest, taskDigest)
	assert.Nil(t, err)
}

func TestGetTaskDigest_Unequal(t *testing.T) {
	compiledTaskRequest := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Id:   &taskIdentifier,
			Type: "foo type",
		},
	}
	changedTaskDigest, err := GetTaskDigest(context.Background(), compiledTaskRequest)
	assert.Nil(t, err)
	assert.NotEqual(t, compiledTaskDigest, changedTaskDigest)
}

func TestGetWorkflowDigest(t *testing.T) {
	compiledWorkflow, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowDigest, err := GetWorkflowDigest(context.Background(), compiledWorkflow)
	assert.Equal(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)
}

func TestGetWorkflowDigest_Unequal(t *testing.T) {
	workflowWithDifferentNodes, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowWithDifferentNodes.Primary.Template.Nodes = append(
		workflowWithDifferentNodes.Primary.Template.Nodes, &core.Node{
			Id: "unexpected",
		})
	workflowDigest, err := GetWorkflowDigest(context.Background(), workflowWithDifferentNodes)
	assert.NotEqual(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)

	workflowWithDifferentInputs, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowWithDifferentInputs.Primary.Template.Interface.Inputs.Variables = append(workflowWithDifferentInputs.Primary.Template.Interface.Inputs.Variables, &core.VariableMapEntry{
		Name: "unexpected",
		Var:  &core.Variable{},
	})
	workflowDigest, err = GetWorkflowDigest(context.Background(), workflowWithDifferentInputs)
	assert.NotEqual(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)

	workflowWithDifferentOutputs, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowWithDifferentOutputs.Primary.Template.Interface.Outputs.Variables = append(workflowWithDifferentOutputs.Primary.Template.Interface.Outputs.Variables, &core.VariableMapEntry{
		Name: "unexpected",
		Var:  &core.Variable{},
	})
	workflowDigest, err = GetWorkflowDigest(context.Background(), workflowWithDifferentOutputs)
	assert.NotEqual(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)
}
