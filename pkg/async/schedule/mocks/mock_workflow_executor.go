// Mock implementation of a WorkflowExecutor for use in tests.
package mocks

import "context"

type MockWorkflowExecutor struct {
	runFunc  func()
	stopFunc func() error
}

func (e *MockWorkflowExecutor) SetRunFunc(runFunc func()) {
	e.runFunc = runFunc
}

func (e *MockWorkflowExecutor) Run(ctx context.Context) {
	if e.runFunc != nil {
		e.runFunc()
	}
}

func (e *MockWorkflowExecutor) SetStopFunc(stopFunc func() error) {
	e.stopFunc = stopFunc
}

func (e *MockWorkflowExecutor) Stop(ctx context.Context) error {
	if e.stopFunc != nil {
		return e.stopFunc()
	}
	return nil
}
