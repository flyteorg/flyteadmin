package common

import (
	"fmt"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"regexp"
	"strconv"
)

type IOType string

const (
	UndefinedIo IOType = ""
	INPUT              = "i"
	OUTPUT             = "o"
	DECK               = "d"
)

func ParseFlyteURL(flyteURL string) (core.NodeExecutionIdentifier, *int, IOType, error) {
	// flyteURL is of the form flyte://v1/project/domain/execution_id/node_id/attempt/[iod]
	// where i stands for inputs.pb o for outputs.pb and d for the flyte deck
	// If the retry attempt is missing, the io requested is assumed to be for the node instead of the task execution
	zero := 0
	re, err := regexp.Compile("flyte://v1/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)(?:/([0-9]+))?/([iod])")
	if err != nil {
		return core.NodeExecutionIdentifier{}, &zero, UndefinedIo, err
	}
	re.MatchString(flyteURL)
	matches := re.FindStringSubmatch(flyteURL)
	if len(matches) != 7 && len(matches) != 6 {
		return core.NodeExecutionIdentifier{}, &zero, UndefinedIo, fmt.Errorf("failed to parse flyte url, only %d matches found", len(matches))
	}
	proj := matches[1]
	domain := matches[2]
	executionID := matches[3]
	nodeID := matches[4]
	var attempt *int // nil means node execution, not a task execution
	if len(matches) == 7 && matches[5] != "" {
		a, err := strconv.Atoi(matches[5])
		if err != nil {
			return core.NodeExecutionIdentifier{}, &zero, UndefinedIo, fmt.Errorf("failed to parse attempt, %s", err)
		}
		attempt = &a
	}
	var ioType IOType
	switch matches[len(matches)-1] {
	case "i":
		ioType = INPUT
	case "o":
		ioType = OUTPUT
	case "d":
		ioType = DECK
	}

	return core.NodeExecutionIdentifier{
		NodeId: nodeID,
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: proj,
			Domain:  domain,
			Name:    executionID,
		},
	}, attempt, ioType, nil
}

func FlyteURLsFromNodeExecutionID(nodeExecutionID core.NodeExecutionIdentifier, deck bool) *admin.FlyteURLs {
	base := fmt.Sprintf("flyte://v1/%s/%s/%s/%s", nodeExecutionID.ExecutionId.Project,
		nodeExecutionID.ExecutionId.Domain, nodeExecutionID.ExecutionId.Name, nodeExecutionID.NodeId)

	res := &admin.FlyteURLs{
		Inputs:  fmt.Sprintf("%s/%s", base, INPUT),
		Outputs: fmt.Sprintf("%s/%s", base, OUTPUT),
	}
	if deck {
		res.Deck = fmt.Sprintf("%s/%s", base, DECK)
	}
	return res
}

func FlyteURLsFromTaskExecutionID(taskExecutionID core.TaskExecutionIdentifier, deck bool) *admin.FlyteURLs {
	base := fmt.Sprintf("flyte://v1/%s/%s/%s/%s/%s", taskExecutionID.NodeExecutionId.ExecutionId.Project,
		taskExecutionID.NodeExecutionId.ExecutionId.Domain, taskExecutionID.NodeExecutionId.ExecutionId.Name, taskExecutionID.NodeExecutionId.NodeId, strconv.Itoa(int(taskExecutionID.RetryAttempt)))

	res := &admin.FlyteURLs{
		Inputs:  fmt.Sprintf("%s/%s", base, INPUT),
		Outputs: fmt.Sprintf("%s/%s", base, OUTPUT),
	}
	if deck {
		res.Deck = fmt.Sprintf("%s/%s", base, DECK)
	}
	return res
}
