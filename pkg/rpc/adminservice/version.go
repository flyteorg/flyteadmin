package adminservice

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func (m *AdminService) GetVersion(ctx context.Context, versionRequest *admin.GetVersionRequest) (*admin.GetVersionResponse, error) {

	response, err := m.VersionManager.GetVersion(ctx, versionRequest)
	if err != nil {
		return nil, err
	}
	return response, nil
}
