package plugins

import (
	"context"
	"testing"
	"time"

	auth "github.com/flyteorg/flyteadmin/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rlStore := newRateLimitStore(1, 1, time.Second)
	assert.NotNil(t, rlStore)
}

func TestLimiterAllow(t *testing.T) {
	rlStore := newRateLimitStore(1, 1, time.Second)
	assert.NoError(t, rlStore.Allow("hello"))
	assert.Error(t, rlStore.Allow("hello"))
	time.Sleep(time.Second)
	assert.NoError(t, rlStore.Allow("hello"))
}

func TestLimiterAllowBurst(t *testing.T) {
	rlStore := newRateLimitStore(1, 2, time.Second)
	assert.NoError(t, rlStore.Allow("hello"))
	assert.NoError(t, rlStore.Allow("hello"))
	assert.Error(t, rlStore.Allow("hello"))
	assert.NoError(t, rlStore.Allow("world"))
}

func TestLimiterClean(t *testing.T) {
	rlStore := newRateLimitStore(1, 1, time.Second)
	assert.NoError(t, rlStore.Allow("hello"))
	assert.Error(t, rlStore.Allow("hello"))
	time.Sleep(time.Second)
	rlStore.clean()
	assert.NoError(t, rlStore.Allow("hello"))
}

func TestLimiterAllowOnMultipleRequests(t *testing.T) {
	rlStore := newRateLimitStore(1, 1, time.Second)
	assert.NoError(t, rlStore.Allow("a"))
	assert.NoError(t, rlStore.Allow("b"))
	assert.NoError(t, rlStore.Allow("c"))
	assert.Error(t, rlStore.Allow("a"))
	assert.Error(t, rlStore.Allow("b"))

	time.Sleep(time.Second)

	assert.NoError(t, rlStore.Allow("a"))
	assert.Error(t, rlStore.Allow("a"))
	assert.NoError(t, rlStore.Allow("b"))
	assert.Error(t, rlStore.Allow("b"))
	assert.NoError(t, rlStore.Allow("c"))
}

func TestRateLimiterLimitPass(t *testing.T) {
	rateLimit := NewRateLimiter(1, 1, time.Second)
	assert.NotNil(t, rateLimit)

	identityCtx, err := auth.NewIdentityContext("audience", "user1", "app1", time.Now(), nil, nil, nil)
	assert.NoError(t, err)

	ctx := context.WithValue(context.TODO(), auth.ContextKeyIdentityContext, identityCtx)
	err = rateLimit.Limit(ctx)
	assert.NoError(t, err)

}

func TestRateLimiterLimitStop(t *testing.T) {
	rateLimit := NewRateLimiter(1, 1, time.Second)
	assert.NotNil(t, rateLimit)

	identityCtx, err := auth.NewIdentityContext("audience", "user1", "app1", time.Now(), nil, nil, nil)
	assert.NoError(t, err)
	ctx := context.WithValue(context.TODO(), auth.ContextKeyIdentityContext, identityCtx)
	err = rateLimit.Limit(ctx)
	assert.NoError(t, err)

	err = rateLimit.Limit(ctx)
	assert.Error(t, err)

}

func TestRateLimiterLimitWithoutUserIdentity(t *testing.T) {
	rateLimit := NewRateLimiter(1, 1, time.Second)
	assert.NotNil(t, rateLimit)

	ctx := context.TODO()

	err := rateLimit.Limit(ctx)
	assert.Error(t, err)
}
