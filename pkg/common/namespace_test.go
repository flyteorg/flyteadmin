package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNamespaceName(t *testing.T) {
	testCases := []struct {
		mapping NamespaceMapping
		project string
		domain  string
		want    string
	}{
		{NmProjectDomain, "project", "production", "project-production"},
		{20 /*Dummy enum value that is not supported*/, "project", "development", "project-development"},
		{NmDomain, "project", "production", "production"},
		{NmProject, "project", "production", "project"},
	}

	for _, tc := range testCases {
		got := GetNamespaceName(tc.mapping, tc.project, tc.domain)
		assert.Equal(t, got, tc.want)
	}
}
