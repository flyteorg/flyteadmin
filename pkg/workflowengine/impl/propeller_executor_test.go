package impl

import (
	"context"
	"fmt"
	"testing"
	"time"

	interfaces2 "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	cluster_mock "github.com/flyteorg/flyteadmin/pkg/executioncluster/mocks"
	"github.com/flyteorg/flyteadmin/pkg/runtime"

	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/flyteorg/flytestdlib/promutils"

	"errors"

	flyte_admin_error "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteclient "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned"
	v1alpha12 "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var fakeFlyteWF = FakeFlyteWorkflowV1alpha1{}
var scope = promutils.NewTestScope()
var propellerTestMetrics = newPropellerMetrics(scope)
var config = runtime.NewConfigurationProvider().NamespaceMappingConfiguration()

var roleNameKey = "iam.amazonaws.com/role"
var clusterName = "C1"

var acceptedAt = time.Now()

func getFlytePropellerForTest(execCluster interfaces2.ClusterInterface, builder *FlyteWorkflowBuilderTest) *FlytePropeller {
	return &FlytePropeller{
		executionCluster: execCluster,
		builder:          builder,
		roleNameKey:      roleNameKey,
		metrics:          propellerTestMetrics,
		config:           config,
	}
}

type buildFlyteWorkflowFunc func(
	wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
	namespace string) (*v1alpha1.FlyteWorkflow, error)
type FlyteWorkflowBuilderTest struct {
	buildCallback buildFlyteWorkflowFunc
}

func (b *FlyteWorkflowBuilderTest) BuildFlyteWorkflow(
	wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
	namespace string) (*v1alpha1.FlyteWorkflow, error) {
	if b.buildCallback != nil {
		return b.buildCallback(wfClosure, inputs, executionID, namespace)
	}
	return &v1alpha1.FlyteWorkflow{}, nil
}

type createCallback func(*v1alpha1.FlyteWorkflow, v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error)
type deleteCallback func(name string, options *v1.DeleteOptions) error
type FakeFlyteWorkflow struct {
	v1alpha12.FlyteWorkflowInterface
	createCallback createCallback
	deleteCallback deleteCallback
}

func (b *FakeFlyteWorkflow) Create(ctx context.Context, wf *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
	if b.createCallback != nil {
		return b.createCallback(wf, opts)
	}
	return nil, nil
}

func (b *FakeFlyteWorkflow) Delete(ctx context.Context, name string, options v1.DeleteOptions) error {
	if b.deleteCallback != nil {
		return b.deleteCallback(name, &options)
	}
	return nil
}

type flyteWorkflowsCallback func(string) v1alpha12.FlyteWorkflowInterface

type FakeFlyteWorkflowV1alpha1 struct {
	v1alpha12.FlyteworkflowV1alpha1Interface
	flyteWorkflowsCallback flyteWorkflowsCallback
}

func (b *FakeFlyteWorkflowV1alpha1) FlyteWorkflows(namespace string) v1alpha12.FlyteWorkflowInterface {
	if b.flyteWorkflowsCallback != nil {
		return b.flyteWorkflowsCallback(namespace)
	}
	return &FakeFlyteWorkflow{}
}

type FakeK8FlyteClient struct {
	flyteclient.Interface
	ID string
}

func (b *FakeK8FlyteClient) FlyteworkflowV1alpha1() v1alpha12.FlyteworkflowV1alpha1Interface {
	return &fakeFlyteWF
}

func getFakeExecutionCluster() interfaces2.ClusterInterface {
	fakeCluster := cluster_mock.MockCluster{}
	fakeCluster.SetGetTargetCallback(func(ctx context.Context, spec *executioncluster.ExecutionTargetSpec) (target *executioncluster.ExecutionTarget, e error) {
		return &executioncluster.ExecutionTarget{
			ID:          "C1",
			FlyteClient: &FakeK8FlyteClient{},
		}, nil
	})
	return &fakeCluster
}

func TestExecuteWorkflowHappyCase(t *testing.T) {
	cluster := cluster_mock.MockCluster{}
	execID := core.WorkflowExecutionIdentifier{
		Project: "p",
		Domain:  "d",
		Name:    "n",
	}
	cluster.SetGetTargetCallback(func(ctx context.Context, spec *executioncluster.ExecutionTargetSpec) (target *executioncluster.ExecutionTarget, e error) {
		assert.Equal(t, execID.Name, spec.ExecutionID)
		return &executioncluster.ExecutionTarget{
			ID:          "C1",
			FlyteClient: &FakeK8FlyteClient{},
		}, nil
	})
	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
			assert.EqualValues(t, map[string]string{
				"customlabel": "labelval",
			}, workflow.Labels)
			expectedAnnotations := map[string]string{
				"iam.amazonaws.com/role": "role-1",
				"customannotation":       "annotationval",
			}
			assert.EqualValues(t, expectedAnnotations, workflow.Annotations)

			assert.EqualValues(t, map[string]v1alpha1.TaskPluginOverride{
				"python": {
					PluginIDs:             []string{"plugin a"},
					MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
				},
			}, workflow.ExecutionConfig.TaskPluginImpls)
			assert.Empty(t, opts)
			return nil, nil
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(&cluster, &FlyteWorkflowBuilderTest{})

	execInfo, err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInput{
			ExecutionID: &execID,
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "lp-p",
					Domain:  "lp-d",
				},
				Spec: &admin.LaunchPlanSpec{
					Role: "role-1",
					WorkflowId: &core.Identifier{
						Name: "wf",
					},
				},
			},
			AcceptedAt: acceptedAt,
			Labels: map[string]string{
				"customlabel": "labelval",
			},
			Annotations: map[string]string{
				"customannotation": "annotationval",
			},
			TaskPluginOverrides: []*admin.PluginOverride{
				{
					TaskType:              "python",
					PluginId:              []string{"plugin a"},
					MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
				},
			},
		})
	assert.Nil(t, err)
	assert.NotNil(t, execInfo)
	assert.Equal(t, clusterName, execInfo.Cluster)
}

func TestExecuteWorkflowCallFailed(t *testing.T) {
	cluster := getFakeExecutionCluster()
	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
			expectedAnnotations := map[string]string{
				"iam.amazonaws.com/role": "role-1",
			}
			assert.EqualValues(t, expectedAnnotations, workflow.Annotations)
			assert.Empty(t, opts)
			return nil, errors.New("call failed")
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(cluster, &FlyteWorkflowBuilderTest{})

	execInfo, err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInput{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
				Spec: &admin.LaunchPlanSpec{
					Role: "role-1",
					WorkflowId: &core.Identifier{
						Name: "wf",
					},
				},
			},
			AcceptedAt: acceptedAt,
		})

	assert.NotNil(t, err)
	assert.Nil(t, execInfo)
	assert.EqualError(t, err, "failed to create workflow in propeller call failed")
	assert.Equal(t, codes.Internal, err.(flyte_admin_error.FlyteAdminError).Code())

}

func TestExecuteWorkflowAlreadyExistsNoError(t *testing.T) {
	cluster := getFakeExecutionCluster()
	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
			expectedAnnotations := map[string]string{
				"iam.amazonaws.com/role": "role-1",
			}
			assert.EqualValues(t, expectedAnnotations, workflow.Annotations)
			assert.Empty(t, opts)
			return nil, k8_api_err.NewAlreadyExists(schema.GroupResource{}, "")
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(cluster, &FlyteWorkflowBuilderTest{})

	execInfo, err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInput{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
				Spec: &admin.LaunchPlanSpec{
					Role: "role-1",
					WorkflowId: &core.Identifier{
						Name: "wf",
					},
				},
			},
			AcceptedAt: acceptedAt,
		})

	assert.Nil(t, err)
	assert.NotNil(t, execInfo)
	assert.Equal(t, clusterName, execInfo.Cluster)
}

func TestExecuteWorkflowBuildFailed(t *testing.T) {
	cluster := getFakeExecutionCluster()
	builder := FlyteWorkflowBuilderTest{}
	builder.buildCallback = func(wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap,
		executionID *core.WorkflowExecutionIdentifier, namespace string) (*v1alpha1.FlyteWorkflow, error) {
		return nil, flyte_admin_error.NewFlyteAdminError(codes.Aborted, "abort")
	}
	propeller := getFlytePropellerForTest(cluster, &builder)
	identifier := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}
	execInfo, err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInput{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{
						Id: &identifier,
					},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
			},
			AcceptedAt: acceptedAt,
		})
	assert.NotNil(t, err)
	assert.EqualError(t, err, fmt.Sprintf("failed to build the workflow [%+v] abort", &identifier))
	assert.Equal(t, codes.Internal, err.(flyte_admin_error.FlyteAdminError).Code())
	assert.Nil(t, execInfo)
}

func TestExecuteWorkflowRoleKeyNotRequired(t *testing.T) {
	cluster := getFakeExecutionCluster()
	builder := FlyteWorkflowBuilderTest{}

	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
			return nil, nil
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(cluster, &builder)
	execInfo, err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInput{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
				Spec: &admin.LaunchPlanSpec{
					WorkflowId: &core.Identifier{
						Name: "wf",
					},
				},
			},
			AcceptedAt: acceptedAt,
		})
	assert.Nil(t, err)
	assert.NotNil(t, execInfo)
}

func TestTerminateExecution(t *testing.T) {
	cluster := getFakeExecutionCluster()
	target, err := cluster.GetTarget(context.Background(), nil)
	assert.Nil(t, err)
	target.ID = "C2"
	builder := FlyteWorkflowBuilderTest{}

	fakeFlyteWorkflow := FakeFlyteWorkflow{
		deleteCallback: func(name string, options *v1.DeleteOptions) error {
			assert.Equal(t, "n", name)
			return errors.New("expected error")
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(cluster, &builder)

	err = propeller.TerminateWorkflowExecution(context.Background(), interfaces.TerminateWorkflowInput{
		ExecutionID: &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		},
		Cluster: "C2",
	})
	assert.EqualError(t, err,
		"failed to terminate execution: project:\"p\" domain:\"d\" name:\"n\"  with err expected error")
}

func TestAddPermissions(t *testing.T) {
	cluster := getFakeExecutionCluster()
	propeller := getFlytePropellerForTest(cluster, &FlyteWorkflowBuilderTest{})
	flyteWf := v1alpha1.FlyteWorkflow{}
	propeller.addPermissions(admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			Auth: &admin.Auth{
				AssumableIamRole: "rollie-pollie",
			},
			Role: "ignore-me",
		},
	}, &flyteWf)
	assert.EqualValues(t, flyteWf.Annotations, map[string]string{
		roleNameKey: "rollie-pollie",
	})

	flyteWf = v1alpha1.FlyteWorkflow{}
	propeller.addPermissions(admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			Role: "rollie-pollie",
		},
	}, &flyteWf)
	assert.EqualValues(t, flyteWf.Annotations, map[string]string{
		roleNameKey: "rollie-pollie",
	})

	flyteWf = v1alpha1.FlyteWorkflow{}
	propeller.addPermissions(admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			Auth: &admin.Auth{
				KubernetesServiceAccount: "service-account",
				AssumableIamRole:         "rollie-pollie",
			},
		},
	}, &flyteWf)
	assert.Equal(t, "service-account", flyteWf.ServiceAccountName)
	assert.EqualValues(t, flyteWf.Annotations, map[string]string{
		roleNameKey: "rollie-pollie",
	})
}
