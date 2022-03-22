package implementations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var testCloudEventPublisher pubsubtest.TestPublisher
var mockCloudEventPublisher pubsub.Publisher = &testCloudEventPublisher
var mockPubSubSender = &PubSubSender{Pub: mockCloudEventPublisher}

// This method should be invoked before every test around Publisher.
func initializeCloudEventPublisher() {
	testCloudEventPublisher.Published = nil
	testCloudEventPublisher.GivenError = nil
	testCloudEventPublisher.FoundError = nil
}

func TestNewCloudEventsPublisher_EventTypes(t *testing.T) {
	{
		tests := []struct {
			name            string
			eventTypes      []string
			events          []proto.Message
			shouldSendEvent []bool
			expectedSendCnt int
		}{
			{"eventTypes as workflow,node", []string{"workflow", "node"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, true, false},
				2},
			{"eventTypes as workflow,task", []string{"workflow", "task"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, false, true},
				2},
			{"eventTypes as workflow,task", []string{"node", "task"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{false, true, true},
				2},
			{"eventTypes as task", []string{"task"},
				[]proto.Message{taskRequest},
				[]bool{true},
				1},
			{"eventTypes as node", []string{"node"},
				[]proto.Message{nodeRequest},
				[]bool{true},
				1},
			{"eventTypes as workflow", []string{"workflow"},
				[]proto.Message{workflowRequest},
				[]bool{true},
				1},
			{"eventTypes as workflow", []string{"workflow"},
				[]proto.Message{nodeRequest, taskRequest},
				[]bool{false, false},
				0},
			{"eventTypes as task", []string{"task"},
				[]proto.Message{workflowRequest, nodeRequest},
				[]bool{false, false},
				0},
			{"eventTypes as node", []string{"node"},
				[]proto.Message{workflowRequest, taskRequest},
				[]bool{false, false},
				0},
			{"eventTypes as all", []string{"all"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, true, true},
				3},
			{"eventTypes as *", []string{"*"},
				[]proto.Message{workflowRequest, nodeRequest, taskRequest},
				[]bool{true, true, true},
				3},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				initializeCloudEventPublisher()
				var currentEventPublisher = NewCloudEventsPublisher(mockPubSubSender, promutils.NewTestScope(), test.eventTypes)
				var cnt = 0
				for id, event := range test.events {
					assert.Nil(t, currentEventPublisher.Publish(context.Background(), proto.MessageName(event),
						event))
					if test.shouldSendEvent[id] {
						assert.Equal(t, proto.MessageName(event), testCloudEventPublisher.Published[cnt].Key)
						body := testCloudEventPublisher.Published[cnt].Body
						cloudevent := cloudevents.NewEvent()
						err := pbcloudevents.Protobuf.Unmarshal(body, &cloudevent)
						assert.Nil(t, err)

						assert.Equal(t, cloudevent.DataContentType(), cloudevents.ApplicationJSON)
						assert.Equal(t, cloudevent.SpecVersion(), cloudevents.VersionV1)
						assert.Equal(t, cloudevent.Type(), proto.MessageName(event))
						assert.Equal(t, cloudevent.Source(), cloudEventSource)
						assert.Equal(t, cloudevent.Extensions(), map[string]interface{}{jsonSchemaURL: "https://github.com/pingsutw/flyteidl/blob/cloudevent2/jsonschema/workflow_execution.json"})

						e, err := json.Marshal(event)
						assert.Nil(t, err)
						assert.Equal(t, cloudevent.Data(), e)
						cnt++
					}
				}
				assert.Equal(t, test.expectedSendCnt, len(testCloudEventPublisher.Published))
			})
		}
	}
}

func TestCloudEventPublisher_PublishError(t *testing.T) {
	initializeCloudEventPublisher()
	currentEventPublisher := NewCloudEventsPublisher(mockPubSubSender, promutils.NewTestScope(), []string{"*"})
	var publishError = errors.New("publish() returns an error")
	testCloudEventPublisher.GivenError = publishError
	assert.Equal(t, publishError, currentEventPublisher.Publish(context.Background(),
		proto.MessageName(taskRequest), taskRequest))
}
