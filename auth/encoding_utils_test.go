package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAscii(t *testing.T) {
	assert.Equal(t, "bmls", EncodeBase64([]byte("nil")))
	assert.Equal(t, "w4RwZmVs", EncodeBase64([]byte("Äpfel")))
}

func TestDecodeFromAscii(t *testing.T) {
	assert.Equal(t, []byte("nil"), DecodeFromBase64("bmls"))
	assert.Equal(t, []byte("Äpfel"), DecodeFromBase64("w4RwZmVs"))
}
