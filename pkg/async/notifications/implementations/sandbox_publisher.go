package implementations

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
)

/*
	TODO: Check if SystemMetrics is necessary for the sandbox publisher.
*/

type SandboxPublisher struct{}

var msgChan = make(chan []byte)

//subscriber can get it from this channel
func (p *SandboxPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	// push the marshal message to the queue
	data, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%s] and error: %v", notificationType, msg.String(), err)
	}
	msgChan <- data

	return nil
}

func NewSandboxPublisher() *SandboxPublisher {
	return &SandboxPublisher{}
}
