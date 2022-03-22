package implementations

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"

	"github.com/NYTimes/gizmo/pubsub"
	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
)

const (
	cloudEventSource = "https://github.com/flyteorg/flyteadmin"
	jsonSchemaURL    = "jsonSchemaURL"
)

type Receiver = string

const (
	Kafka Receiver = "Kafka"
)

type CloudEventSender interface {
	Send(ctx context.Context, notificationType string, event cloudevents.Event) error
}

type PubSubSender struct {
	Pub pubsub.Publisher
}

func (s *PubSubSender) Send(ctx context.Context, notificationType string, event cloudevents.Event) error {
	eventByte, err := pbcloudevents.Protobuf.Marshal(&event)
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal cloudevent with error: %v", err)
		return err
	}
	if err := s.Pub.PublishRaw(ctx, notificationType, eventByte); err != nil {
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%s] and error: %v", notificationType, event.String(), err)
		return err
	}

	return nil
}

type KafkaSender struct {
	Client cloudevents.Client
}

func (s *KafkaSender) Send(ctx context.Context, notificationType string, event cloudevents.Event) error {
	if result := s.Client.Send(
		// Set the producer message key
		kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(event.ID())),
		event,
	); cloudevents.IsUndelivered(result) {
		logger.Errorf(ctx, "failed to send: %v", result)
	}
	return nil
}

type CloudEventPublisher struct {
	sender        CloudEventSender
	systemMetrics eventPublisherSystemMetrics
	events        sets.String
}

func (p *CloudEventPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	p.systemMetrics.PublishTotal.Inc()

	if !p.shouldPublishEvent(notificationType) {
		return nil
	}
	logger.Debugf(ctx, "Publishing the following message [%+v]", msg)

	event := cloudevents.NewEvent()
	// CloudEvent specification: https://github.com/cloudevents/spec/blob/v1.0/spec.md#required-attributes
	event.SetType(notificationType)
	event.SetSource(cloudEventSource)
	event.SetID(uuid.New().String())
	event.SetTime(time.Now())
	event.SetExtension(jsonSchemaURL, "https://github.com/pingsutw/flyteidl/blob/cloudevent2/jsonschema/workflow_execution.json")

	if err := event.SetData(cloudevents.ApplicationJSON, &msg); err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to encode data with error: %v", err)
		return err
	}

	if err := p.sender.Send(ctx, notificationType, event); err != nil {
		p.systemMetrics.PublishError.Inc()
		return err
	}
	return nil
}

func (p *CloudEventPublisher) shouldPublishEvent(notificationType string) bool {
	return p.events.Has(notificationType)
}

func NewCloudEventsPublisher(sender CloudEventSender, scope promutils.Scope, eventTypes []string) interfaces.Publisher {
	eventSet := sets.NewString()

	for _, event := range eventTypes {
		if event == AllTypes || event == AllTypesShort {
			for _, e := range supportedEvents {
				eventSet = eventSet.Insert(e)
			}
			break
		}
		if e, found := supportedEvents[event]; found {
			eventSet = eventSet.Insert(e)
		} else {
			logger.Errorf(context.Background(), "Unsupported event type [%s] in the config")
		}
	}

	return &CloudEventPublisher{
		sender:        sender,
		systemMetrics: newEventPublisherSystemMetrics(scope.NewSubScope("cloudevents_publisher")),
		events:        eventSet,
	}
}
