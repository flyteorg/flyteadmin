package impl

import (
	"context"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	adminversion "github.com/flyteorg/flyteadmin/version"
)

type VersionManager struct {
	Version string
	Build string
	BuildTime string
}

func (v *VersionManager) GetVersion(_ context.Context) (*admin.Version, error) {
	return &admin.Version{
		Version : v.Version,
		Build: v.Build,
		BuildTime: v.BuildTime,
	}, nil
}

func NewVersionManager() interfaces.VersionInterface {
	return &VersionManager{
		Version : adminversion.Version,
		Build: adminversion.Build,
		BuildTime: adminversion.BuildTime,
	}
}
