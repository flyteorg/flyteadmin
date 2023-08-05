package implementations

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

var mockSandboxEmailer mocks.MockEmailer

func TestSandboxProcessor_StartProcessing(t *testing.T) {
	msgChan := make(chan []byte, 1)
	testSandboxProcessor := NewSandboxProcessor(msgChan, &mockSandboxEmailer)

	sendEmailValidationFunc := func(ctx context.Context, email admin.EmailMessage) error {
		assert.Equal(t, testEmail.Body, email.Body)
		assert.Equal(t, testEmail.RecipientsEmail, email.RecipientsEmail)
		assert.Equal(t, testEmail.SubjectLine, email.SubjectLine)
		assert.Equal(t, testEmail.SenderEmail, email.SenderEmail)
		return nil
	}
	mockSandboxEmailer.SetSendEmailFunc(sendEmailValidationFunc)
	assert.Nil(t, testSandboxProcessor.(*SandboxProcessor).run())
}
