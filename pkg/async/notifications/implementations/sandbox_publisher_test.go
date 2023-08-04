package implementations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandboxPublisher_Publish(t *testing.T) {
	msgChan := make(chan []byte, 1)
	publisher := NewSandboxPublisher(msgChan)

	err := publisher.Publish(context.Background(), "NOTIFICATION_TYPE", &testEmail)

	assert.NotZero(t, len(msgChan))
	assert.Nil(t, err)
}
