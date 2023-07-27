package implementations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSandboxPublisher_Publish(t *testing.T) {
	// Initialize the sandbox publisher
	publisher := NewSandboxPublisher()

	// Create a channel to communicate failures
	errChan := make(chan string)

	go func() {
		select {
		case <-msgChan:
			// if message received, no need to send an error
		case <-time.After(time.Second * 5):
			errChan <- "No data was received in the channel within the expected time frame"
		}
	}()

	// Run the Publish method
	err := publisher.Publish(context.Background(), "NOTIFICATION_TYPE", &testEmail)

	// Check if there was an error in the goroutine
	select {
	case errMsg := <-errChan:
		t.Fatal(errMsg)
	default:
		// no error from the goroutine
	}

	// Ensure there was no error in the Publish method
	assert.Nil(t, err)
}
