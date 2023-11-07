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
	mutex      *sync.Mutex
}

// LimiterStore stores the access records for each user
type LimiterStore struct {
	// accessPerUser is a synchronized map of userID to accessRecords
	accessPerUser   *sync.Map
	requestPerSec   int
	burstSize       int
	cleanupInterval time.Duration
}

// Allow takes a userID and returns an error if the user has exceeded the rate limit
func (l *LimiterStore) Allow(userID string) error {
	accessRecord, _ := l.accessPerUser.LoadOrStore(userID, &accessRecords{
		limiter:    rate.NewLimiter(rate.Limit(l.requestPerSec), l.burstSize),
		lastAccess: time.Now(),
		mutex:      &sync.Mutex{},
	})
	accessRecord.(*accessRecords).mutex.Lock()
	defer accessRecord.(*accessRecords).mutex.Unlock()

	accessRecord.(*accessRecords).lastAccess = time.Now()
	l.accessPerUser.Store(userID, accessRecord)

	if !accessRecord.(*accessRecords).limiter.Allow() {
		return RateLimitExceeded(fmt.Errorf("rate limit exceeded"))
	}

	return nil
}

// clean removes the access records for users who have not accessed the system for a while
func (l *LimiterStore) clean() {
	l.accessPerUser.Range(func(key, value interface{}) bool {
		value.(*accessRecords).mutex.Lock()
		defer value.(*accessRecords).mutex.Unlock()
		if time.Since(value.(*accessRecords).lastAccess) > l.cleanupInterval {
			l.accessPerUser.Delete(key)
		}
		return true
	})
}

func newRateLimitStore(requestPerSec int, burstSize int, cleanupInterval time.Duration) *LimiterStore {
	l := &LimiterStore{
		accessPerUser:   &sync.Map{},
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

// RateLimiter is a struct that implements the RateLimiter interface from grpc middleware
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
