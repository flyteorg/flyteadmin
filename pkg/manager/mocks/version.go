package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type MockVersionManager struct {
	createWorkflowFunc CreateWorkflowFunc
}

func (r *MockVersionManager) GetVersion(ctx context.Context) (*admin.Version, error) {
	return nil, nil
}
