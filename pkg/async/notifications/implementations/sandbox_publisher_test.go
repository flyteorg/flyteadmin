package implementations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSandboxPublisher_Publish(t *testing.T) {
	publisher := NewSandboxPublisher()

	errChan := make(chan string)

	go func() {
		select {
		case <-msgChan:
			// if message received, no need to send an error
		case <-time.After(time.Second * 5):
			errChan <- "No data was received in the channel within the expected time frame"
		}
	}()

	err := publisher.Publish(context.Background(), "NOTIFICATION_TYPE", &testEmail)

	// Check if there was an error in the goroutine
	select {
	case errMsg := <-errChan:
		t.Fatal(errMsg)
	default:
		// no error from the goroutine
	}

	assert.Nil(t, err)
}
