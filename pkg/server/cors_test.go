package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPatterns(t *testing.T) {
	pattern := GetGlobPattern()
	x, err := pattern.Match([]string{"api", "v1", "executions", "flytekit", "production"}, "")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(x))
}
