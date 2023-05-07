package plugins

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Second)
	assert.NotNil(t, rl)
}

func TestLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Second)
	assert.NoError(t, rl.Allow("hello"))
	// assert error type is RateLimitError
	assert.Error(t, rl.Allow("hello"))
	time.Sleep(time.Second)
	assert.NoError(t, rl.Allow("hello"))
}

func TestLimiter_AllowBurst(t *testing.T) {
	rl := NewRateLimiter(1, 2, time.Second)
	assert.NoError(t, rl.Allow("hello"))
	assert.NoError(t, rl.Allow("hello"))
	assert.Error(t, rl.Allow("hello"))
	assert.NoError(t, rl.Allow("world"))
}

func TestLimiter_Clean(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Second)
	assert.NoError(t, rl.Allow("hello"))
	assert.Error(t, rl.Allow("hello"))
	time.Sleep(time.Second)
	rl.clean()
	assert.NoError(t, rl.Allow("hello"))
}

func TestLimiter_AllowOnMultipleRequests(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Second)
	assert.NoError(t, rl.Allow("a"))
	assert.NoError(t, rl.Allow("b"))
	assert.NoError(t, rl.Allow("c"))
	assert.Error(t, rl.Allow("a"))
	assert.Error(t, rl.Allow("b"))

	time.Sleep(time.Second)

	assert.NoError(t, rl.Allow("a"))
	assert.Error(t, rl.Allow("a"))
	assert.NoError(t, rl.Allow("b"))
	assert.Error(t, rl.Allow("b"))
	assert.NoError(t, rl.Allow("c"))
}
