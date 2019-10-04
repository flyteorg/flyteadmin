package auth

import (
	"context"
	"encoding/base64"
	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"testing"
)

// This function can also be called locally to generate new keys
func TestSecureCookieLifecycle(t *testing.T)  {
	hashKey := securecookie.GenerateRandomKey(64)
	assert.True(t, base64.RawStdEncoding.EncodeToString(hashKey) != "")

	blockKey := securecookie.GenerateRandomKey(32)
	assert.True(t, base64.RawStdEncoding.EncodeToString(blockKey) != "")

	cookie, err := NewSecureCookie("choc", "chip", hashKey, blockKey)
	assert.NoError(t, err)

	value, err := ReadSecureCookie(context.Background(), cookie, hashKey, blockKey)
	assert.NoError(t, err)
	assert.Equal(t, "chip", value)
}

