package data

import (
	"testing"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/data/implementations"
	"github.com/stretchr/testify/assert"
)

func TestGetRemoteDataHandlerForGCP(t *testing.T) {
	c := RemoteDataHandlerConfig{
		CloudProvider:            common.GCP,
		SignedURLDurationMinutes: 1,
		SigningPrincipal:         "principal@example.com",
	}
	h := GetRemoteDataHandler(c)
	assert.NotNil(t, h)
	assert.IsType(t, &implementations.GCPRemoteURL{}, h.GetRemoteURLInterface())
}
