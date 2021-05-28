package implementations

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestAddresses(t *testing.T) {
	addresses := []string{"alice@example.com", "bob@example.com"}
	sgAddresses := getEmailAddresses(addresses)
	assert.Equal(t, sgAddresses[0].Address, "alice@example.com")
	assert.Equal(t, sgAddresses[1].Address, "bob@example.com")
}

func TestGetEmail(t *testing.T) {
	emailNotification := admin.EmailMessage{
		SubjectLine: "Notice: Execution \"name\" has succeeded in \"domain\".",
		SenderEmail: "no-reply@example.com",
		RecipientsEmail: []string{
			"my@example.com",
			"john@example.com",
		},
		Body: "Execution \"name\" has succeeded in \"domain\". View details at " +
			"<a href=\"https://example.com/executions/T/B/D\">" +
			"https://example.com/executions/T/B/D</a>.",
	}

	sgEmail := getSendgridEmail(emailNotification)
	assert.Equal(t, `Notice: Execution "name" has succeeded in "domain".`, sgEmail.Personalizations[0].Subject)
	assert.Equal(t, "john@example.com", sgEmail.Personalizations[0].To[1].Address)
	assert.Equal(t, `Execution "name" has succeeded in "domain". View details at <a href="https://example.com/executions/T/B/D">https://example.com/executions/T/B/D</a>.`, sgEmail.Content[0].Value)
}

func TestCreateEmailer(t *testing.T) {
	cfg := getNotificationsConfig()
	cfg.NotificationsEmailerConfig.EmailerConfig.ApiKeyEnvVar = "sendgrid_api_key"

	emailer := NewSendGridEmailer(cfg, promutils.NewTestScope())
	assert.NotNil(t, emailer)
}
