package implementations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCPSplitURI(t *testing.T) {
	remoteURL := GCPRemoteURL{}
	gcsObject, err := remoteURL.splitURI(context.Background(), "gs://i/am/valid")
	assert.Nil(t, err)
	assert.Equal(t, "i", gcsObject.bucket)
	assert.Equal(t, "am/valid", gcsObject.object)
}

func TestGCPSplitURI_InvalidScheme(t *testing.T) {
	remoteURL := GCPRemoteURL{}
	_, err := remoteURL.splitURI(context.Background(), "azure://i/am/invalid")
	assert.NotNil(t, err)
}

func TestGCPSplitURI_InvalidDataReference(t *testing.T) {
	remoteURL := GCPRemoteURL{}
	_, err := remoteURL.splitURI(context.Background(), "s3://invalid\\")
	assert.NotNil(t, err)
}
