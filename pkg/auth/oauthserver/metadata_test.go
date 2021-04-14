package oauthserver

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetJSONWebKeys(t *testing.T) {
	newpriv, err := rsa.GenerateMultiPrimeKey(rand.Reader, 4, 128)
	if err != nil {
		t.Errorf("failed to generate key")
	}
	oldpriv, err := rsa.GenerateMultiPrimeKey(rand.Reader, 4, 128)
	if err != nil {
		t.Errorf("failed to generate key")
	}
	newKey := newpriv.PublicKey
	oldKey := oldpriv.PublicKey
	publicKeys := []rsa.PublicKey{newKey, oldKey}
	keyset, err := getJSONWebKeys(publicKeys)
	assert.Nil(t, err)
	assert.NotNil(t, keyset)
	oldJwkKey, exists := keyset.Get(1)
	assert.True(t, exists)
	oldpublicKey, exists := oldJwkKey.Get(KeyMetadataPublicCert)
	op, ok := oldpublicKey.(*rsa.PublicKey)
	assert.True(t, ok)
	assert.Equal(t, &oldKey, op)
	newJwkKey, exists := keyset.Get(0)
	assert.True(t, exists)
	newpublicKey, exists := newJwkKey.Get(KeyMetadataPublicCert)
	np, ok := newpublicKey.(*rsa.PublicKey)
	assert.True(t, ok)
	assert.NotEqual(t, np, op)
	assert.Equal(t, &newKey, np)
}
