package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNamespaceName(t *testing.T) {
	testCases := []struct {
		mapping string
		project string
		domain  string
		want    string
	}{
		{"", "project", "production", "project-production"},
		{"bad-value", "project", "development", "project-development"},
		{"domain", "project", "production", "production"},
	}

	for _, tc := range testCases {
		got := GetNamespaceName(tc.mapping, tc.project, tc.domain)
		assert.Equal(t, got, tc.want)
	}
}
