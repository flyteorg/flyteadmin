package plugins

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	auth "github.com/flyteorg/flyteadmin/auth"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RateLimitExceeded error

// accessRecords stores the rate limiter and the last access time
type accessRecords struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// LimiterStore stores the access records for each user
type LimiterStore struct {
	// accessPerUser is a synchronized map of userID to accessRecords
	accessPerUser   map[string]*accessRecords
	mutex           *sync.Mutex
	requestPerSec   int
	burstSize       int
	cleanupInterval time.Duration
}

// Allow takes a userID and returns an error if the user has exceeded the rate limit
func (l *LimiterStore) Allow(userID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if _, ok := l.accessPerUser[userID]; !ok {
		l.accessPerUser[userID] = &accessRecords{
			lastAccess: time.Now(),
			limiter:    rate.NewLimiter(rate.Limit(l.requestPerSec), l.burstSize),
		}
	}
	l.accessPerUser[userID].lastAccess = time.Now()

	if !l.accessPerUser[userID].limiter.Allow() {
		return RateLimitExceeded(fmt.Errorf("rate limit exceeded"))
	}

	return nil
}

func (l *LimiterStore) clean() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for userID, accessRecord := range l.accessPerUser {
		if time.Since(accessRecord.lastAccess) > l.cleanupInterval {
			delete(l.accessPerUser, userID)
		}
	}
}

func newRateLimitStore(requestPerSec int, burstSize int, cleanupInterval time.Duration) *LimiterStore {
	l := &LimiterStore{
		accessPerUser:   make(map[string]*accessRecords),
		mutex:           &sync.Mutex{},
		requestPerSec:   requestPerSec,
		burstSize:       burstSize,
		cleanupInterval: cleanupInterval,
	}

	go func() {
		for {
			time.Sleep(l.cleanupInterval)
			l.clean()
		}
	}()

	return l
}

type RateLimiter struct {
	limiter *LimiterStore
}

func (r *RateLimiter) Limit(ctx context.Context) error {
	IdenCtx := auth.IdentityContextFromContext(ctx)
	if IdenCtx.IsEmpty() {
		return errors.New("no identity context found")
	}
	userID := IdenCtx.UserID()
	if err := r.limiter.Allow(userID); err != nil {
		return err
	}
	return nil
}

func NewRateLimiter(requestPerSec int, burstSize int, cleanupInterval time.Duration) *RateLimiter {
	limiter := newRateLimitStore(requestPerSec, burstSize, cleanupInterval)
	return &RateLimiter{limiter: limiter}
}

func RateLimiteInterceptor(limiter RateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp interface{}, err error) {
		if err := limiter.Limit(ctx); err != nil {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}
