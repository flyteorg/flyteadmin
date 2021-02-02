package implementations

import (
	"context"
	"strings"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/lyft/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

type eventPublisherSystemMetrics struct {
	Scope        promutils.Scope
	PublishTotal prometheus.Counter
	PublishError prometheus.Counter
}

// TODO: Add a counter that encompasses the publisher stats grouped by project and domain.
type EventPublisher struct {
	pub           pubsub.Publisher
	systemMetrics eventPublisherSystemMetrics
	events        []string
}

func getSupportedEvents() map[string]string {
	supportedEvents := make(map[string]string)
	var taskExecutionReq admin.TaskExecutionEventRequest
	supportedEvents["task"] = proto.MessageName(&taskExecutionReq)
	var nodeExecutionReq admin.NodeExecutionEventRequest
	supportedEvents["node"] = proto.MessageName(&nodeExecutionReq)
	var workflowExecutionReq admin.WorkflowExecutionEventRequest
	supportedEvents["workflow"] = proto.MessageName(&workflowExecutionReq)
	return supportedEvents
}

// The key is the notification type as defined as an enum.
func (p *EventPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	p.systemMetrics.PublishTotal.Inc()
	logger.Debugf(ctx, "Publishing the following message [%s]", msg.String())

	if !p.shouldPublishEvent(notificationType) {
		return nil
	}

	err := p.pub.Publish(ctx, notificationType, msg)
	if err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%s] and error: %v", notificationType, msg.String(), err)
	}
	return err
}

func (p *EventPublisher) shouldPublishEvent(notificationType string) bool {
	for _, e := range p.events {
		if e == notificationType {
			return true
		}
	}
	return false
}

func newEventPublisherSystemMetrics(scope promutils.Scope) eventPublisherSystemMetrics {
	return eventPublisherSystemMetrics{
		Scope:        scope,
		PublishTotal: scope.MustNewCounter("event_publish_total", "overall count of publish messages"),
		PublishError: scope.MustNewCounter("event_publish_errors", "count of publish errors"),
	}
}

func NewEventsPublisher(pub pubsub.Publisher, scope promutils.Scope, eventTypes string) interfaces.Publisher {
	supportedEvents := getSupportedEvents()

	var eventList = make([]string, 0)
	if strings.Contains(eventTypes, "*") || strings.Contains(eventTypes, "all") {
		for _, e := range supportedEvents {
			eventList = append(eventList, e)
		}
	} else {
		events := strings.Split(eventTypes, ",")
		for _, event := range events {
			if e, found := supportedEvents[event]; found {
				eventList = append(eventList, e)
			}
		}
	}

	return &EventPublisher{
		pub:           pub,
		systemMetrics: newEventPublisherSystemMetrics(scope.NewSubScope("events_publisher")),
		events:        eventList,
	}
}
