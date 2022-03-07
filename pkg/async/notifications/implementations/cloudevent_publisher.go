package implementations

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/google/uuid"
	"github.com/invopop/jsonschema"

	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
)

const (
	cloudEventSource = "https://github.com/flyteorg/flyteadmin"
	jsonSchema       = "jsonschema"
)

type CloudEventPublisher struct {
	pub           pubsub.Publisher
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
	reflector := jsonschema.Reflector{ExpandedStruct: true}
	schema, err := json.Marshal(reflector.Reflect(msg))
	if err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to marshal cloudevent JsonSchema: %v", err)
		return err
	}
	event.SetExtension(jsonSchema, string(schema))

	if err := event.SetData(cloudevents.ApplicationJSON, &msg); err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to encode data with error: %v", err)
		return err
	}

	eventByte, err := pbcloudevents.Protobuf.Marshal(&event)
	if err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to marshal cloudevent with error: %v", err)
		return err
	}
	if err := p.pub.PublishRaw(ctx, notificationType, eventByte); err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%s] and error: %v", notificationType, msg.String(), err)
		return err
	}

	return nil
}

func (p *CloudEventPublisher) shouldPublishEvent(notificationType string) bool {
	return p.events.Has(notificationType)
}

func NewCloudEventsPublisher(pub pubsub.Publisher, scope promutils.Scope, eventTypes []string) interfaces.Publisher {
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
		pub:           pub,
		systemMetrics: newEventPublisherSystemMetrics(scope.NewSubScope("cloudevents_publisher")),
		events:        eventSet,
	}
}
