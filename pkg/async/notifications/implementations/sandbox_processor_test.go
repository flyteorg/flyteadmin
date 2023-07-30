package implementations

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var mockSandboxEmailer mocks.MockEmailer

func TestSandboxProcessor_UnmarshalMessage(t *testing.T) {
	var emailMessage admin.EmailMessage
	err := proto.Unmarshal(msg, &emailMessage)
	assert.Nil(t, err)

	assert.Equal(t, emailMessage.Body, testEmail.Body)
	assert.Equal(t, emailMessage.RecipientsEmail, testEmail.RecipientsEmail)
	assert.Equal(t, emailMessage.SubjectLine, testEmail.SubjectLine)
	assert.Equal(t, emailMessage.SenderEmail, testEmail.SenderEmail)
}

func TestSandboxProcessor_StartProcessing(t *testing.T) {

	testSandboxProcessor := NewSandboxProcessor(&mockSandboxEmailer)

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
