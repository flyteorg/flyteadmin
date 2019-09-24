package implementations

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/lyft/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/lyft/flyteadmin/pkg/errors"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type emailMetrics struct {
	Scope       promutils.Scope
	SendSuccess prometheus.Counter
	SendError   prometheus.Counter
	SendTotal   prometheus.Counter
}

func newEmailMetrics(scope promutils.Scope) emailMetrics {
	return emailMetrics{
		Scope:       scope,
		SendSuccess: scope.MustNewCounter("send_success", "Number of successful emails sent via Emailer."),
		SendError:   scope.MustNewCounter("send_error", "Number of errors when sending email via Emailer"),
		SendTotal:   scope.MustNewCounter("send_total", "Total number of emails attempted to be sent"),
	}
}

type AwsEmailer struct {
	config        runtimeInterfaces.NotificationsConfig
	systemMetrics emailMetrics
	awsEmail      sesiface.SESAPI
}

func (e *AwsEmailer) SendEmail(ctx context.Context, email admin.EmailMessage) error {
	var toAddress []*string
	for _, toEmail := range email.RecipientsEmail {
		toAddress = append(toAddress, &toEmail)
	}

	emailInput := ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: toAddress,
		},
		// Currently use the senderEmail specified apart of the Emailer instead of the body.
		// Once a more generic way of setting the emailNotification is defined, remove this
		// workaround and defer back to email.SenderEmail
		Source: &email.SenderEmail,
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Data: &email.Body,
				},
			},
			Subject: &ses.Content{
				Data: &email.SubjectLine,
			},
		},
	}

	_, err := e.awsEmail.SendEmail(&emailInput)
	e.systemMetrics.SendTotal.Inc()

	if err != nil {
		// TODO: If we see a certain set of AWS errors consistently, we can break the errors down based on type.
		logger.Errorf(ctx, "error in sending email [%s] via ses mailer with err: %s", email.String(), err)
		e.systemMetrics.SendError.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "errors were seen while sending emails")
	}

	e.systemMetrics.SendSuccess.Inc()
	return nil
}

func NewAwsEmailer(config runtimeInterfaces.NotificationsConfig, scope promutils.Scope, awsEmail sesiface.SESAPI) interfaces.Emailer {
	return &AwsEmailer{
		config:        config,
		systemMetrics: newEmailMetrics(scope.NewSubScope("aws_ses")),
		awsEmail:      awsEmail,
	}
}
