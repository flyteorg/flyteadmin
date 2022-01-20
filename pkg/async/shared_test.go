package async

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	attemptsRecorded := 0
	err := Retry(3, time.Millisecond, func() error {
		if attemptsRecorded == 3 {
			return nil
		}
		attemptsRecorded++
		return errors.New("foo")
	})
	assert.Nil(t, err)
}

func TestRetry_RetriesExhausted(t *testing.T) {
	attemptsRecorded := 0
	err := Retry(2, time.Millisecond, func() error {
		if attemptsRecorded == 3 {
			return nil
		}
		attemptsRecorded++
		return errors.New("foo")
	})
	assert.EqualError(t, err, "foo")
}

func TestRetryDoNotRetryOnNonRetryableExceptions(t *testing.T) {
	attemptsRecorded := 0
	err := RetryOnSpecificErrors(3, time.Millisecond, func() error {
		attemptsRecorded++
		return errors.New("foo")
	}, func(err error) bool {
		// Function will cause retry on no errors
		return false
	})
	assert.EqualValues(t, attemptsRecorded, 1)
	assert.EqualError(t, err, "foo")
}

func TestRetryOnRetryableExceptions(t *testing.T) {
	attemptsRecorded := 0
	err := RetryOnSpecificErrors(3, time.Millisecond, func() error {
		attemptsRecorded++
		return errors.New("foo")
	}, func(err error) bool {
		// Function will cause retry on all errors
		return true
	})
	assert.EqualValues(t, attemptsRecorded, 4)
	assert.EqualError(t, err, "foo")
}
