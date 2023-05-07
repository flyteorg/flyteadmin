package plugins

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitError error

// define a struct that contains a map of rate limiters, and a time stamp of last access and a mutex to protect the map
type accessRecords struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

type Limiter struct {
	accessPerUser   map[string]*accessRecords
	mutex           *sync.Mutex
	requestPerSec   int
	burstSize       int
	cleanupInterval time.Duration
}

// define a function named Allow that takes userID and returns RateLimitError
// the function check if the user is in the map, if not, create a new accessRecords for the user
// then it check if the user can access the resource, if not, return RateLimitError
func (l *Limiter) Allow(userID string) RateLimitError {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if _, ok := l.accessPerUser[userID]; !ok {
		l.accessPerUser[userID] = &accessRecords{
			lastAccess: time.Now(),
			limiter:    rate.NewLimiter(rate.Limit(l.requestPerSec), l.burstSize),
		}
	}

	if !l.accessPerUser[userID].limiter.Allow() {
		return RateLimitError(fmt.Errorf("rate limit exceeded"))
	}

	return nil
}

func (l *Limiter) clean() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for userID, accessRecord := range l.accessPerUser {
		if time.Since(accessRecord.lastAccess) > l.cleanupInterval {
			delete(l.accessPerUser, userID)
		}
	}
}

func NewRateLimiter(requestPerSec int, burstSize int, cleanupInterval time.Duration) *Limiter {
	l := &Limiter{
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
