package interfaces

// FlyteK8sWorkflowExecutorRegistry is a singleton source of which K8sWorkflowExecutor implementation to use for
// creating and deleting Flyte workflow CRD objects.
type FlyteK8sWorkflowExecutorRegistry interface {
	Register(executor K8sWorkflowExecutor)
	RegisterDefault(executor K8sWorkflowExecutor)
	GetExecutor() K8sWorkflowExecutor
}
