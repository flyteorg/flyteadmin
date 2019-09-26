package executioncluster

import (
	"github.com/lyft/flyteadmin/pkg/flytek8s"
	runtime "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	flyteclient "github.com/lyft/flytepropeller/pkg/client/clientset/versioned"
	"github.com/lyft/flytestdlib/promutils"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExecutionTargetSpec struct {
	TargetID string
}

type ExecutionTarget struct {
	ID          string
	FlyteClient flyteclient.Interface
	Client      client.Client
}

func getRestClientFromKubeConfig(scope promutils.Scope, kubeConfiguration *rest.Config) (*flyteclient.Clientset, error) {
	fc, err := flyteclient.NewForConfig(kubeConfiguration)
	if err != nil {
		scope.MustNewCounter(
			"flyteclient_initialization_error",
			"count of errors encountered initializing a flyte client from kube config").Inc()
		return nil, err
	}
	return fc, nil
}

func NewExecutionTarget(scope promutils.Scope, k8sCluster runtime.ClusterConfig) (*ExecutionTarget, error) {
	kubeConf, err := flytek8s.GetRestClientConfigForCluster(k8sCluster)
	if err != nil {
		return nil, err
	}
	flyteClient, err := getRestClientFromKubeConfig(scope, kubeConf)
	if err != nil {
		return nil, err
	}
	client, err := client.New(kubeConf, client.Options{})
	if err != nil {
		return nil, err
	}
	return &ExecutionTarget{
		FlyteClient: flyteClient,
		Client:      client,
		ID:          k8sCluster.Name,
	}, nil
}
