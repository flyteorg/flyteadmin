package impl

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/pkg/errors"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/golang/protobuf/ptypes/timestamp"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type metrics struct {
	Scope promutils.Scope
	//Set   labeled.Counter
}

type MetricsManager struct {
	db                   repoInterfaces.Repository
	workflowManager      interfaces.WorkflowInterface
	executionManager     interfaces.ExecutionInterface
	nodeExecutionManager interfaces.NodeExecutionInterface
	taskExecutionManager interfaces.TaskExecutionInterface
	metrics              metrics
}

func createCategoricalSpan(startTime, endTime *timestamp.Timestamp, category admin.CategoricalSpanInfo_Category) *admin.Span {
	return &admin.Span{
		StartTime: startTime,
		EndTime: endTime,
		Info: &admin.Span_Category{
			Category: &admin.CategoricalSpanInfo{
				Category:    category,
			},
		},
	}
}

func (m *MetricsManager) getLatestUpstreamNodeExecution(ctx context.Context, nodeId string,
	upstreamNodeIds map[string]*core.ConnectionSet_IdList, nodeExecutions map[string]*admin.NodeExecution) (*admin.NodeExecution, error) {

	var nodeExecution *admin.NodeExecution
	var latestUpstreamUpdatedAt = time.Unix(0, 0)
	if connectionSet, exists := upstreamNodeIds[nodeId]; exists {
		for _, upstreamNodeId := range connectionSet.Ids {
			upstreamNodeExecution := nodeExecutions[upstreamNodeId]

			t := upstreamNodeExecution.Closure.UpdatedAt.AsTime()
			if t.After(latestUpstreamUpdatedAt) {
				nodeExecution = upstreamNodeExecution
				latestUpstreamUpdatedAt = t
			}
		}
	}

	return nodeExecution, nil
}

func (m *MetricsManager) getNodeExecutions(ctx context.Context, request admin.NodeExecutionListRequest) (map[string]*admin.NodeExecution, error) {
	nodeExecutions := make(map[string]*admin.NodeExecution)
	for {
		response, err := m.nodeExecutionManager.ListNodeExecutions(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, nodeExecution := range response.NodeExecutions {
			nodeExecutions[nodeExecution.Metadata.SpecNodeId] = nodeExecution
		}

		if len(response.NodeExecutions) < int(request.Limit) {
			break
		}

		request.Token = response.Token
	}

	return nodeExecutions, nil
}

func (m *MetricsManager) getTaskExecutions(ctx context.Context, request admin.TaskExecutionListRequest) ([]*admin.TaskExecution, error) {
	taskExecutions := make([]*admin.TaskExecution, 0)
	for {
		response, err := m.taskExecutionManager.ListTaskExecutions(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, taskExecution := range response.TaskExecutions {
			taskExecutions = append(taskExecutions, taskExecution)
		}

		if len(response.TaskExecutions) < int(request.Limit) {
			break
		}

		request.Token = response.Token
	}

	return taskExecutions, nil
}

func (m *MetricsManager) parseDynamicNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution,
	taskExecutions []*admin.TaskExecution, nodeExecutions map[string]*admin.NodeExecution, spans *[]*admin.Span, depth int) error {

	getDataRequest := admin.NodeExecutionGetDataRequest{Id: nodeExecution.Id}
	nodeExecutionData, err := m.nodeExecutionManager.GetNodeExecutionData(ctx, getDataRequest)
	if err != nil {
		return err
	}

	// frontend overhead
	*spans = append(*spans, createCategoricalSpan(nodeExecution.Closure.CreatedAt,
		taskExecutions[0].Closure.CreatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	// task execution(s)
	parseTaskExecutions(taskExecutions, spans, depth)

	// between task execution(s) and node execution(s) overhead
	startNode := nodeExecutions["start-node"]
	*spans = append(*spans, createCategoricalSpan(taskExecutions[len(taskExecutions)-1].Closure.UpdatedAt,
		startNode.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	// node execution(s)
	if err := m.parseNodeExecutions(ctx, nodeExecutions, nodeExecutionData.DynamicWorkflow.CompiledWorkflow, spans, depth); err != nil {
		return err
	}

	// backened overhead
	latestUpstreamNode, err := m.getLatestUpstreamNodeExecution(ctx, "end-node",
		nodeExecutionData.DynamicWorkflow.CompiledWorkflow.Primary.Connections.Upstream, nodeExecutions)
	if err != nil {
		return err
	}

	*spans = append(*spans, createCategoricalSpan(latestUpstreamNode.Closure.UpdatedAt,
		nodeExecution.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	return nil
}

func (m *MetricsManager) parseExecution(ctx context.Context, execution *admin.Execution, depth int) (*admin.Span, error) {
	referenceSpan := &admin.ReferenceSpanInfo{
		Id: &admin.ReferenceSpanInfo_WorkflowId{
			WorkflowId: execution.Id,
		},
	}

	if depth != 0 {
		spans := make([]*admin.Span, 0)

		// retrieve workflow and node executions
		workflowRequest := admin.ObjectGetRequest{Id: execution.Closure.WorkflowId}
		workflow, err := m.workflowManager.GetWorkflow(ctx, workflowRequest)
		if err != nil {
			return nil, err
		}

		nodeListRequest := admin.NodeExecutionListRequest{
			WorkflowExecutionId: execution.Id,
			Limit: 20, // TODO @hamersaw - parameterize?
		}
		nodeExecutions, err := m.getNodeExecutions(ctx, nodeListRequest)
		if err != nil {
			return nil, err
		}

		// compute frontend overhead
		startNode := nodeExecutions["start-node"]
		spans = append(spans, createCategoricalSpan(execution.Closure.CreatedAt,
			startNode.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))
		
		// iterate over nodes and compute overhead
		if err := m.parseNodeExecutions(ctx, nodeExecutions, workflow.Closure.CompiledWorkflow, &spans, depth-1); err != nil {
			return nil, err
		}

		// compute backend overhead
		latestUpstreamNode, err := m.getLatestUpstreamNodeExecution(ctx, "end-node",
			workflow.Closure.CompiledWorkflow.Primary.Connections.Upstream, nodeExecutions)
		if err != nil {
			return nil, err
		}

		spans = append(spans, createCategoricalSpan(latestUpstreamNode.Closure.UpdatedAt,
			execution.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

		referenceSpan.Spans = spans
	}

	return &admin.Span{
		StartTime: execution.Closure.CreatedAt,
		EndTime: execution.Closure.UpdatedAt,
		Info: &admin.Span_Reference{
			Reference: referenceSpan,
		},
	}, nil
}

func (m *MetricsManager) parseLaunchPlanNodeExecution(ctx context.Context,
	nodeExecution *admin.NodeExecution, spans *[]*admin.Span, depth int) error {

	// retrieve execution
	workflowNode := nodeExecution.Closure.GetWorkflowNodeMetadata()

	executionRequest := admin.WorkflowExecutionGetRequest{Id: workflowNode.ExecutionId}
	execution, err := m.executionManager.GetExecution(ctx, executionRequest)
	if err != nil {
		return err
	}

	// frontend overhead
	*spans = append(*spans, createCategoricalSpan(nodeExecution.Closure.CreatedAt,
		execution.Closure.CreatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	// execution
	span, err := m.parseExecution(ctx, execution, depth)
	if err != nil {
		return err
	}

	*spans = append(*spans, span)

	// backend overhead
	*spans = append(*spans, createCategoricalSpan(execution.Closure.UpdatedAt,
		nodeExecution.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	return nil
}

func (m *MetricsManager) parseNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution, node *core.Node, depth int) (*admin.Span, error) {
	referenceSpan := &admin.ReferenceSpanInfo{
		Id: &admin.ReferenceSpanInfo_NodeId{
			NodeId: nodeExecution.Id,
		},
	}

	if depth != 0 {
		spans := make([]*admin.Span, 0)

		// TODO @hamersaw - move these into the node parsing functions
		//   no need to get node executions for a taskNode / etc
		// retrieve task and node executions
		taskListRequest := admin.TaskExecutionListRequest{
			NodeExecutionId: nodeExecution.Id,
			Limit: 20, // TODO @hamersaw - parameterize?
		}
		taskExecutions, err := m.getTaskExecutions(ctx, taskListRequest)
		if err != nil {
			return nil, err
		}

		nodeListRequest := admin.NodeExecutionListRequest{
			WorkflowExecutionId: nodeExecution.Id.ExecutionId,
			Limit: 20, // TODO @hamersaw - parameterize?
			UniqueParentId: nodeExecution.Id.NodeId,
		}
		nodeExecutions, err := m.getNodeExecutions(ctx, nodeListRequest)
		if err != nil {
			return nil, err
		}

		// parse node
		switch target := node.Target.(type) {
			case *core.Node_BranchNode:
			case *core.Node_GateNode:
			case *core.Node_TaskNode:
				if nodeExecution.Metadata.IsParentNode {
					// handle dynamic node
					if err := m.parseDynamicNodeExecution(ctx, nodeExecution, taskExecutions, nodeExecutions, &spans, depth-1); err != nil {
						return nil, err
					}
				} else {
					// handle task node
					m.parseTaskNodeExecution(ctx, nodeExecution, taskExecutions, &spans, depth-1)
				}
			case *core.Node_WorkflowNode:
				switch workflow := target.WorkflowNode.Reference.(type) {
					case *core.WorkflowNode_LaunchplanRef:
						// handle launch plan
						if err := m.parseLaunchPlanNodeExecution(ctx, nodeExecution, &spans, depth-1); err != nil {
							return nil, err
						}
					case *core.WorkflowNode_SubWorkflowRef:
						// handle subworkflow
						if err := m.parseSubworkflowNodeExecution(ctx, nodeExecution, workflow.SubWorkflowRef, nodeExecutions, &spans, depth-1); err != nil {
							return nil, err
						}
				}
			default:
				fmt.Printf("unsupported node type %+v\n", target)
		}

		referenceSpan.Spans = spans
	}

	return &admin.Span{
		StartTime: nodeExecution.Closure.CreatedAt,
		EndTime: nodeExecution.Closure.UpdatedAt,
		Info: &admin.Span_Reference{
			Reference: referenceSpan,
		},
	}, nil
}

func (m *MetricsManager) parseNodeExecutions(ctx context.Context, nodeExecutions map[string]*admin.NodeExecution,
	compiledWorkflowClosure *core.CompiledWorkflowClosure, spans *[]*admin.Span, depth int) error {

	// sort node executions
	sortedNodeExecutions := make([]*admin.NodeExecution, 0, len(nodeExecutions))
    for _, nodeExecution := range nodeExecutions {
        sortedNodeExecutions = append(sortedNodeExecutions, nodeExecution)
    }
	sort.Slice(sortedNodeExecutions, func(i, j int) bool {
		x := sortedNodeExecutions[i].Closure.CreatedAt.AsTime()
		y := sortedNodeExecutions[j].Closure.CreatedAt.AsTime()
		return x.Before(y)
	})

	// iterate over sorted node executions
	for _, nodeExecution := range sortedNodeExecutions {
		specNodeId := nodeExecution.Metadata.SpecNodeId
		if specNodeId == "start-node" || specNodeId == "end-node" {
			continue
		}

		fmt.Printf("HAMERSAW - parsing node %s\n", specNodeId)

		// identify subworkflow from node id
		var node *core.Node
		for _, n := range compiledWorkflowClosure.Primary.Template.Nodes {
			if n.Id == specNodeId{
				node = n
			}
		}

		if node == nil {
			return errors.New("failed to identify workflow node") // TODO @hamersaw - do gooder
		}

		// parse node execution
		nodeExecutionSpan, err := m.parseNodeExecution(ctx, nodeExecution, node, depth)
		if err != nil {
			return err
		}

		// prepend nodeExecution spans with NODE_TRANSITION time
		if referenceSpan, ok := nodeExecutionSpan.Info.(*admin.Span_Reference); ok {
			latestUpstreamNode, err := m.getLatestUpstreamNodeExecution(ctx, specNodeId,
				compiledWorkflowClosure.Primary.Connections.Upstream, nodeExecutions)
			if err != nil {
				return err
			}

			// TODO @hamersaw - check if latestUpstreamNode is nil
			referenceSpan.Reference.Spans = append([]*admin.Span{createCategoricalSpan(latestUpstreamNode.Closure.UpdatedAt,
				nodeExecution.Closure.CreatedAt, admin.CategoricalSpanInfo_NODE_TRANSITION)}, referenceSpan.Reference.Spans...)
		}

		*spans = append(*spans, nodeExecutionSpan)
	}

	return nil
}

func (m *MetricsManager) parseSubworkflowNodeExecution(ctx context.Context,
	nodeExecution *admin.NodeExecution, identifier *core.Identifier, nodeExecutions map[string]*admin.NodeExecution, spans *[]*admin.Span, depth int) error {

	// retrieve workflow
	workflowRequest := admin.ObjectGetRequest{Id: identifier}
	workflow, err := m.workflowManager.GetWorkflow(ctx, workflowRequest)
	if err != nil {
		return err
	}

	// frontend overhead
	startNode := nodeExecutions["start-node"]
	*spans = append(*spans, createCategoricalSpan(nodeExecution.Closure.CreatedAt,
		startNode.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	// node execution(s)
	if err := m.parseNodeExecutions(ctx, nodeExecutions, workflow.Closure.CompiledWorkflow, spans, depth); err != nil {
		return err
	}

	// backened overhead
	latestUpstreamNode, err := m.getLatestUpstreamNodeExecution(ctx, "end-node",
		workflow.Closure.CompiledWorkflow.Primary.Connections.Upstream, nodeExecutions)
	if err != nil {
		return err
	}

	*spans = append(*spans, createCategoricalSpan(latestUpstreamNode.Closure.UpdatedAt,
		nodeExecution.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	return nil
}

func parseTaskExecution(taskExecution *admin.TaskExecution) *admin.Span {
	spans := make([]*admin.Span, 0)
	spans = append(spans, createCategoricalSpan(taskExecution.Closure.CreatedAt,
		taskExecution.Closure.StartedAt, admin.CategoricalSpanInfo_PLUGIN_OVERHEAD))

	taskEndTime := timestamppb.New(taskExecution.Closure.StartedAt.AsTime().Add(taskExecution.Closure.Duration.AsDuration()))
	spans = append(spans, createCategoricalSpan(taskExecution.Closure.StartedAt,
		taskEndTime, admin.CategoricalSpanInfo_PLUGIN_EXECUTION))

	spans = append(spans, createCategoricalSpan(taskEndTime,
		taskExecution.Closure.UpdatedAt, admin.CategoricalSpanInfo_PLUGIN_OVERHEAD))

	return &admin.Span{
		StartTime: taskExecution.Closure.CreatedAt,
		EndTime:   taskExecution.Closure.UpdatedAt,
		Info:      &admin.Span_Reference{
			Reference: &admin.ReferenceSpanInfo{
				Id: &admin.ReferenceSpanInfo_TaskId{
					TaskId: taskExecution.Id,
				},
				Spans: spans,
			},
		},
	}
}

func parseTaskExecutions(taskExecutions []*admin.TaskExecution, spans *[]*admin.Span, depth int) {
	// sort task executions
	sort.Slice(taskExecutions, func(i, j int) bool {
		x := taskExecutions[i].Closure.CreatedAt.AsTime()
		y := taskExecutions[j].Closure.CreatedAt.AsTime()
		return x.Before(y)
	})

	// iterate over task executions
	for index, taskExecution := range taskExecutions {
		if index > 0 {
			*spans = append(*spans, createCategoricalSpan(taskExecutions[index-1].Closure.UpdatedAt,
				taskExecution.Closure.CreatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))
		}

		if depth != 0 {
			*spans = append(*spans, parseTaskExecution(taskExecution))
		}
	}
}

func (m *MetricsManager) parseTaskNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution,
	taskExecutions []*admin.TaskExecution, spans *[]*admin.Span, depth int) {

	*spans = append(*spans, createCategoricalSpan(nodeExecution.Closure.CreatedAt,
		taskExecutions[0].Closure.CreatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	parseTaskExecutions(taskExecutions, spans, depth)

	*spans = append(*spans, createCategoricalSpan(taskExecutions[len(taskExecutions)-1].Closure.UpdatedAt,
		nodeExecution.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))
}

// TODO @hamersaw - docs
func (m *MetricsManager) GetExecutionMetrics(ctx context.Context,
	request admin.WorkflowExecutionGetMetricsRequest) (*admin.WorkflowExecutionGetMetricsResponse, error) {
	
	// retrieve workflow execution
	executionRequest := admin.WorkflowExecutionGetRequest{Id: request.Id}
	execution, err := m.executionManager.GetExecution(ctx, executionRequest)
	if err != nil {
		return nil, err
	}

	span, err := m.parseExecution(ctx, execution, int(request.Depth))
	if err != nil {
		return nil, err
	}

	return &admin.WorkflowExecutionGetMetricsResponse{Span: span}, nil
}

// TODO @hamersaw docs
func (m *MetricsManager) GetNodeExecutionMetrics(ctx context.Context,
	request admin.NodeExecutionGetMetricsRequest) (*admin.NodeExecutionGetMetricsResponse, error) {

	// retrieve node execution
	nodeRequest := admin.NodeExecutionGetRequest{Id: request.Id}
	nodeExecution, err := m.nodeExecutionManager.GetNodeExecution(ctx, nodeRequest)
	if err != nil {
		return nil, err
	}

	span, err := m.parseNodeExecution(ctx, nodeExecution, nil, int(request.Depth)) // TODO @hamersaw can NOT pass nil for Node - FIX IMMEDIATELY
	if err != nil {
		return nil, err
	}

	return &admin.NodeExecutionGetMetricsResponse{Span: span}, nil
}

func NewMetricsManager(
	db repoInterfaces.Repository,
	workflowManager interfaces.WorkflowInterface,
	executionManager interfaces.ExecutionInterface,
	nodeExecutionManager interfaces.NodeExecutionInterface,
	taskExecutionManager interfaces.TaskExecutionInterface,
	scope promutils.Scope) interfaces.MetricsInterface {
	metrics := metrics{
		Scope: scope,
		//Set:   labeled.NewCounter("num_set", "count of set metricss", scope),
	}

	return &MetricsManager{
		db:                   db,
		workflowManager:      workflowManager,
		executionManager:     executionManager,
		nodeExecutionManager: nodeExecutionManager,
		taskExecutionManager: taskExecutionManager,
		metrics:              metrics,
	}
}
