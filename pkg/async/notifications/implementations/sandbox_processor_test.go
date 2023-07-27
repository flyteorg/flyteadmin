package implementations

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestSandboxProcessor_StartProcessing(t *testing.T) {
	// Mock a message
	var emailMessage admin.EmailMessage
	err := proto.Unmarshal(msg, &emailMessage)
	assert.Nil(t, err)

	assert.Equal(t, emailMessage.Body, testEmail.Body)
	assert.Equal(t, emailMessage.RecipientsEmail, testEmail.RecipientsEmail)
	assert.Equal(t, emailMessage.SubjectLine, testEmail.SubjectLine)
	assert.Equal(t, emailMessage.SenderEmail, testEmail.SenderEmail)
}
