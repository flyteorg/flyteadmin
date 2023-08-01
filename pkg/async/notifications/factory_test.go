package notifications

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/implementations"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

var (
	scope               = promutils.NewScope("test_sandbox_processor")
	notificationsConfig = runtimeInterfaces.NotificationsConfig{
		Type: "sandbox",
	}
)

func TestGetEmailer(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.NotificationsConfig{
		NotificationsEmailerConfig: runtimeInterfaces.NotificationsEmailerConfig{
			EmailerConfig: runtimeInterfaces.EmailServerConfig{
				ServiceName: "unsupported",
			},
		},
	}

	GetEmailer(cfg, promutils.NewTestScope())

	// shouldn't reach here
	t.Errorf("did not panic")
}

func TestNewNotificationsProcessor(t *testing.T) {
	testSandboxProcessor := NewNotificationsProcessor(notificationsConfig, scope)
	assert.IsType(t, testSandboxProcessor, &implementations.SandboxProcessor{})
}

func TestNewNotificationPublisher(t *testing.T) {
	testSandboxPublisher := NewNotificationsPublisher(notificationsConfig, scope)
	assert.IsType(t, testSandboxPublisher, &implementations.SandboxPublisher{})
}
