package cloudevent

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/implementations"

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
	jsonSchemaURLKey = "jsonschemaurl"
	jsonSchemaURL    = "https://github.com/flyteorg/flyteidl/blob/cloudevent2/jsonschema/workflow_execution.json"
)

type Receiver = string

const (
	Kafka Receiver = "Kafka"
)

// Sender Defines the interface for sending cloudevents.
type Sender interface {
	// Send a cloud event to other services (pub/sub or Kafka)
	Send(ctx context.Context, notificationType string, event cloudevents.Event) error
}

// PubSubSender Implementation of Sender
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

// KafkaSender Implementation of Sender
type KafkaSender struct {
	Client cloudevents.Client
}

func (s *KafkaSender) Send(ctx context.Context, notificationType string, event cloudevents.Event) error {
	if result := s.Client.Send(
		// Set the producer message key
		kafka_sarama.WithMessageKey(context.Background(), sarama.StringEncoder(event.ID())),
		event,
	); cloudevents.IsUndelivered(result) {
		return fmt.Errorf("failed to send event: %v", result)
	}
	return nil
}

// Publisher This event publisher acts to asynchronously publish workflow execution events.
type Publisher struct {
	sender        Sender
	systemMetrics implementations.EventPublisherSystemMetrics
	events        sets.String
}

func (p *Publisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	if !p.shouldPublishEvent(notificationType) {
		return nil
	}
	logger.Debugf(ctx, "Publishing the following message [%+v]", msg)
	p.systemMetrics.PublishTotal.Inc()

	event := cloudevents.NewEvent()
	// CloudEvent specification: https://github.com/cloudevents/spec/blob/v1.0/spec.md#required-attributes
	event.SetType(notificationType)
	event.SetSource(cloudEventSource)
	event.SetID(uuid.New().String())
	event.SetTime(time.Now())
	event.SetExtension(jsonSchemaURLKey, jsonSchemaURL)

	if err := event.SetData(cloudevents.ApplicationJSON, &msg); err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to encode data with error: %v", err)
		return err
	}

	if err := p.sender.Send(ctx, notificationType, event); err != nil {
		p.systemMetrics.PublishError.Inc()
		return err
	}
	p.systemMetrics.PublishSuccess.Inc()
	return nil
}

func (p *Publisher) shouldPublishEvent(notificationType string) bool {
	return p.events.Has(notificationType)
}

func NewCloudEventsPublisher(sender Sender, scope promutils.Scope, eventTypes []string) interfaces.Publisher {
	eventSet := sets.NewString()

	for _, event := range eventTypes {
		if event == implementations.AllTypes || event == implementations.AllTypesShort {
			for _, e := range implementations.SupportedEvents {
				eventSet = eventSet.Insert(e)
			}
			break
		}
		if e, found := implementations.SupportedEvents[event]; found {
			eventSet = eventSet.Insert(e)
		} else {
			panic(fmt.Errorf("unsupported event type [%s] in the config", event))
		}
	}

	return &Publisher{
		sender:        sender,
		systemMetrics: implementations.NewEventPublisherSystemMetrics(scope.NewSubScope("cloudevents_publisher")),
		events:        eventSet,
	}
}
