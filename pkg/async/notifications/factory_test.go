package notifications

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/flyteorg/flyteadmin/pkg/async/cloudevent"
	"github.com/flyteorg/flyteadmin/pkg/common"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
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

func TestGetCloudEventPublisher(t *testing.T) {
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}

	t.Run("disable cloud event publisher", func(t *testing.T) {
		cfg.Enable = false
		assert.NotNil(t, NewCloudEventsPublisher(cfg, promutils.NewTestScope()))
	})

	t.Run("local publisher", func(t *testing.T) {
		cfg.Enable = false
		cfg.Type = common.Local
		assert.NotNil(t, NewCloudEventsPublisher(cfg, promutils.NewTestScope()))
	})

	t.Run("aws config", func(t *testing.T) {
		cfg.AWSConfig = runtimeInterfaces.AWSConfig{Region: "us-east-1"}
		cfg.Type = common.AWS
		assert.NotNil(t, NewCloudEventsPublisher(cfg, promutils.NewTestScope()))
	})

	t.Run("gcp config", func(t *testing.T) {
		cfg.GCPConfig = runtimeInterfaces.GCPConfig{ProjectID: "project"}
		cfg.Type = common.GCP
		assert.NotNil(t, NewCloudEventsPublisher(cfg, promutils.NewTestScope()))
	})

	t.Run("kafka config", func(t *testing.T) {
		version, err := sarama.ParseKafkaVersion("0.8.2.0")
		assert.Nil(t, err)
		cfg.Type = cloudevent.Kafka
		cfg.KafkaConfig = runtimeInterfaces.KafkaConfig{Version: version}
		assert.NotNil(t, NewCloudEventsPublisher(cfg, promutils.NewTestScope()))
	})
}

func TestInvalidAwsConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  common.AWS,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}
	NewCloudEventsPublisher(cfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}

func TestInvalidGcpConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  common.GCP,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}
	NewCloudEventsPublisher(cfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}

func TestInvalidKafkaConfig(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.CloudEventsConfig{
		Enable:                true,
		Type:                  cloudevent.Kafka,
		EventsPublisherConfig: runtimeInterfaces.EventsPublisherConfig{TopicName: "topic"},
	}
	NewCloudEventsPublisher(cfg, promutils.NewTestScope())
	t.Errorf("did not panic")
}
