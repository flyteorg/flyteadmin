package implementations

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/proto"
)

type SandboxProcessor struct {
	email interfaces.Emailer
	// systemMetrics processorSystemMetrics
}

func (p *SandboxProcessor) StartProcessing() {
	for {
		logger.Warningf(context.Background(), "Starting SandBox notifications processor")
		err := p.run()
		logger.Errorf(context.Background(), "error with running SandBox processor err: [%v] ", err)
		time.Sleep(async.RetryDelay)
	}
}

func (p *SandboxProcessor) run() error {
	var emailMessage admin.EmailMessage

	// use select instead
	select {
	case msg := <-msgChan:
		err := proto.Unmarshal(msg, &emailMessage)
		if err != nil {
			logger.Errorf(context.Background(), "error with unmarshalling message [%v] ", err)
		}
		err = p.email.SendEmail(context.Background(), emailMessage)
	default:
		logger.Debugf(context.Background(), "no message to process")
	}

	return nil

}

func (p *SandboxProcessor) StopProcessing() error {
	logger.Debug(context.Background(), "call to sandbox stop processing.")
	return nil
}

func NewSandboxProcessor(emailer interfaces.Emailer) interfaces.Processor {
	return &SandboxProcessor{
		email: emailer,
		// systemMetrics: newProcessorSystemMetrics(scope.NewSubScope("sandbox_processor")),
	}
}
