package implementations

import (
	"context"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"os"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

type SendgridEmailer struct {
	client        *sendgrid.Client
	systemMetrics emailMetrics
}

func (s SendgridEmailer) SendEmail(ctx context.Context, email admin.EmailMessage) error {
	from := mail.NewEmail("Flyte Notifications", "workflow-notifications@uniondemo.run")
	subject := email.SubjectLine
	to := mail.NewEmail("Sou", "souffle@example.com") // Change to your recipient
	message := mail.NewSingleEmail(from, subject, to, "", email.Body)
	s.systemMetrics.SendTotal.Inc()
	response, err := s.client.Send(message)
	if err != nil {
		logger.Errorf(ctx, "Sendgrid error sending %s", err)
		s.systemMetrics.SendError.Inc()
		return err
	} else {
		s.systemMetrics.SendSuccess.Inc()
		logger.Debugf(ctx, "Sendgrid sent email %s", response.Body)
	}
	return nil
}

func NewSendGridEmailer(config runtimeInterfaces.NotificationsConfig, scope promutils.Scope) interfaces.Emailer {
	return &SendgridEmailer{
		//client:        sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY")),
		client:        sendgrid.NewSendClient(os.Getenv(config.NotificationsEmailerConfig.EmailerConfig.ApiKeyEnvVar)),
		systemMetrics: newEmailMetrics(scope.NewSubScope("sendgrid")),
	}
}
