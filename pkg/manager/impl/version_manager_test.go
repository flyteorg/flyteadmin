package impl

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestVersionManager_GetVersion(t *testing.T) {
	versionManager := NewVersionManager()
	_, err := versionManager.GetVersion(context.Background(), &admin.GetVersionRequest{})
	assert.Nil(t, err)
}
