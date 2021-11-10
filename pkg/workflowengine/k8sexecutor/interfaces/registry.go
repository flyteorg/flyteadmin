package interfaces

// FlyteK8sWorkflowExecutorRegistry is a singleton provider of a K8sWorkflowExecutor implementation to use for
// creating and deleting Flyte workflow CRD objects.
type FlyteK8sWorkflowExecutorRegistry interface {
	// Registers a new K8sWorkflowExecutor to handle creating and aborting Flyte workflow executions.
	Register(executor K8sWorkflowExecutor)
	// Registers the default K8sWorkflowExecutor to handle creating and aborting Flyte workflow executions.
	RegisterDefault(executor K8sWorkflowExecutor)
	// Resolves the definitive K8sWorkflowExecutor implementation to be used for creating and aborting Flyte workflow executions.
	GetExecutor() K8sWorkflowExecutor
}
