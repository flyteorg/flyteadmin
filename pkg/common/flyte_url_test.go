package common

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestParseFlyteUrl(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		ne, attempt, kind, err := ParseFlyteURL("flyte://v1/fs/dev/abc/n0/0/o")
		assert.NoError(t, err)
		assert.Equal(t, 0, *attempt)
		assert.Equal(t, ArtifactTypeO, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))
		ne, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0/i")
		assert.NoError(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeI, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))

		ne, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0/d")
		assert.NoError(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeD, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))

		ne, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0-dn0-9-n0-n0/d")
		assert.NoError(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeD, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0-dn0-9-n0-n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))
	})

	t.Run("invalid", func(t *testing.T) {
		// more than one character
		_, attempt, kind, err := ParseFlyteURL("flyte://v1/fs/dev/abc/n0/0/od")
		assert.Error(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeUndefined, kind)

		_, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0/input")
		assert.Error(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeUndefined, kind)

		// non integer for attempt
		_, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/ab/n0/a/i")
		assert.Error(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeUndefined, kind)
	})
}

func TestFlyteURLsFromNodeExecutionID(t *testing.T) {
	t.Run("with deck", func(t *testing.T) {
		ne := core.NodeExecutionIdentifier{
			NodeId: "n0-dn0-n1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}
		urls := FlyteURLsFromNodeExecutionID(ne, true)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/o", urls.GetOutputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/d", urls.GetDeck())
	})

	t.Run("without deck", func(t *testing.T) {
		ne := core.NodeExecutionIdentifier{
			NodeId: "n0-dn0-n1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}
		urls := FlyteURLsFromNodeExecutionID(ne, false)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/o", urls.GetOutputs())
		assert.Equal(t, "", urls.GetDeck())
	})
}

func TestFlyteURLsFromTaskExecutionID(t *testing.T) {
	t.Run("with deck", func(t *testing.T) {
		te := core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "fs",
				Domain:       "dev",
				Name:         "abc",
				Version:      "v1",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "n0",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "fs",
					Domain:  "dev",
					Name:    "abc",
				},
			},
			RetryAttempt: 1,
		}
		urls := FlyteURLsFromTaskExecutionID(te, true)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/1/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/1/o", urls.GetOutputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/1/d", urls.GetDeck())
	})

	t.Run("without deck", func(t *testing.T) {
		te := core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "fs",
				Domain:       "dev",
				Name:         "abc",
				Version:      "v1",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "n0",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "fs",
					Domain:  "dev",
					Name:    "abc",
				},
			},
		}
		urls := FlyteURLsFromTaskExecutionID(te, false)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/0/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/0/o", urls.GetOutputs())
		assert.Equal(t, "", urls.GetDeck())
	})
}

func TestMatchRegexDirectly(t *testing.T) {
	result := MatchRegex(re, "flyte://v1/fs/dev/abc/n0-dn0-9-n0-n0/i")
	assert.Equal(t, "", result["attempt"])

	result = MatchRegex(re, "flyteff://v2/fs/dfdsaev/abc/n0-dn0-9-n0-n0/i")
	assert.Nil(t, result)
}

func TestDirectRegexMatching(t *testing.T) {
	t.Run("regex with specific output no attempt", func(t *testing.T) {
		matches := MatchRegex(reSpecificOutput, "flyte://v1/fs/dev/abc/n0/o/o0")
		fmt.Println(matches)

	})

	t.Run("regex with specific output no attempt", func(t *testing.T) {
		matches := MatchRegex(reSpecificOutput, "flyte://v1/fs/dev/abc/n0-dn0-9-n0-n0/o/o0")
		fmt.Println(matches)
	})

	t.Run("regex with specific output with attempt", func(t *testing.T) {
		matches := MatchRegex(reSpecificOutput, "flyte://v1/fs/dev/abc/n0-dn0-9-n0-n0/5/o/o0")
		fmt.Println(matches)

		normal := MatchRegex(re, "flyte://v1/fs/dev/abc/n0-dn0-9-n0-n0/5/o/o0")
		assert.Equal(t, 0, len(normal))
	})
}

func TestTryMatches(t *testing.T) {
	t.Run("workflow level", func(t *testing.T) {
		x := tryMatches("fdjskflds")
		assert.Nil(t, x)

		x = tryMatches("flyte://v1/fs/dev/abc/n0/o/o0")
		assert.Equal(t, "o0", x["LiteralName"])

		x = tryMatches("flyte://v1/fs/dev/abc/n0/3/o/o0")
		assert.Equal(t, "fs", x["project"])
		assert.Equal(t, "dev", x["domain"])
		assert.Equal(t, "o0", x["LiteralName"])
		assert.Equal(t, "3", x["attempt"])
		assert.Equal(t, "n0", x["node"])
		assert.Equal(t, "abc", x["exec"])

		x = tryMatches("flyte://v1/fs/dev/abc/n0/3/i")
		assert.Equal(t, "fs", x["project"])
		assert.Equal(t, "dev", x["domain"])
		assert.Equal(t, "", x["LiteralName"])
		assert.Equal(t, "3", x["attempt"])
		assert.Equal(t, "n0", x["node"])
		assert.Equal(t, "abc", x["exec"])

		x = tryMatches("flyte://v1/fs/dev/abc/n0/i")
		assert.Equal(t, "fs", x["project"])
		assert.Equal(t, "dev", x["domain"])
		assert.Equal(t, "", x["LiteralName"])
		assert.Equal(t, "", x["attempt"])
		assert.Equal(t, "n0", x["node"])
		assert.Equal(t, "abc", x["exec"])
	})
}

func TestParseFlyteURLToExecution(t *testing.T) {
	t.Run("node and attempt url with output", func(t *testing.T) {
		x, err := ParseFlyteURLToExecution("flyte://v1/fs/dev/abc/n0/3/o/o0")
		assert.NoError(t, err)
		assert.Nil(t, x.NodeExecID)
		assert.Nil(t, x.PartialTaskExecID.TaskId)
		assert.Equal(t, "fs", x.PartialTaskExecID.NodeExecutionId.ExecutionId.Project)
		assert.Equal(t, "dev", x.PartialTaskExecID.NodeExecutionId.ExecutionId.Domain)
		assert.Equal(t, "abc", x.PartialTaskExecID.NodeExecutionId.ExecutionId.Name)
		assert.Equal(t, "n0", x.PartialTaskExecID.NodeExecutionId.NodeId)
		assert.Equal(t, uint32(3), x.PartialTaskExecID.GetRetryAttempt())
		assert.Equal(t, "o0", x.LiteralName)
	})

	t.Run("node and attempt url no output", func(t *testing.T) {
		x, err := ParseFlyteURLToExecution("flyte://v1/fs/dev/abc/n0/3/o")
		assert.NoError(t, err)
		assert.Nil(t, x.NodeExecID)
		assert.Nil(t, x.PartialTaskExecID.TaskId)
		assert.Equal(t, "fs", x.PartialTaskExecID.NodeExecutionId.ExecutionId.Project)
		assert.Equal(t, "dev", x.PartialTaskExecID.NodeExecutionId.ExecutionId.Domain)
		assert.Equal(t, "abc", x.PartialTaskExecID.NodeExecutionId.ExecutionId.Name)
		assert.Equal(t, "n0", x.PartialTaskExecID.NodeExecutionId.NodeId)
		assert.Equal(t, uint32(3), x.PartialTaskExecID.GetRetryAttempt())
		assert.Equal(t, "", x.LiteralName)
	})

	t.Run("node url with output", func(t *testing.T) {
		x, err := ParseFlyteURLToExecution("flyte://v1/fs/dev/abc/n0/o/o0")
		assert.NoError(t, err)
		assert.NotNil(t, x.NodeExecID)
		assert.Nil(t, x.PartialTaskExecID)
		assert.Equal(t, "fs", x.NodeExecID.ExecutionId.Project)
		assert.Equal(t, "dev", x.NodeExecID.ExecutionId.Domain)
		assert.Equal(t, "abc", x.NodeExecID.ExecutionId.Name)
		assert.Equal(t, "n0", x.NodeExecID.NodeId)
		assert.Equal(t, "o0", x.LiteralName)
	})

	t.Run("node url no output", func(t *testing.T) {
		x, err := ParseFlyteURLToExecution("flyte://v1/fs/dev/abc/n0/o")
		assert.NoError(t, err)
		assert.NotNil(t, x.NodeExecID)
		assert.Nil(t, x.PartialTaskExecID)
		assert.Equal(t, "fs", x.NodeExecID.ExecutionId.Project)
		assert.Equal(t, "dev", x.NodeExecID.ExecutionId.Domain)
		assert.Equal(t, "abc", x.NodeExecID.ExecutionId.Name)
		assert.Equal(t, "n0", x.NodeExecID.NodeId)
		assert.Equal(t, "", x.LiteralName)
	})
}
