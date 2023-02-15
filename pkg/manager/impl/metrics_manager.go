package impl

import (
	"context"
	//"fmt"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

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

func (m *MetricsManager) getLatestUpstreamNodeExecution(ctx context.Context, nodeId string, executionId *core.WorkflowExecutionIdentifier,
	upstreamNodeIds map[string]*core.ConnectionSet_IdList, nodeCache map[string]*admin.NodeExecution) (*admin.NodeExecution, error) {

	var nodeExecution *admin.NodeExecution
	var latestUpstreamUpdatedAt = time.Unix(0, 0)
	if connectionSet, exists := upstreamNodeIds[nodeId]; exists {
		for _, upstreamNodeId := range connectionSet.Ids {
			upstreamNodeExecution, err := m.getNodeExecution(ctx, upstreamNodeId, executionId, nodeCache)
			if err != nil {
				return nil, err // TODO @hamersaw - is this right?
			}

			t := upstreamNodeExecution.Closure.UpdatedAt.AsTime()
			if t.After(latestUpstreamUpdatedAt) {
				nodeExecution = upstreamNodeExecution
				latestUpstreamUpdatedAt = t
			}
		}
	}

	return nodeExecution, nil
}

// TODO @hamersaw - docs
func (m *MetricsManager) getNodeExecution(ctx context.Context, nodeId string, executionId *core.WorkflowExecutionIdentifier,
	nodeCache map[string]*admin.NodeExecution) (*admin.NodeExecution, error) {

	// if node already exists in cache -> use it
	if nodeExecution, exists := nodeCache[nodeId]; exists {
		return nodeExecution, nil
	}

	// retrieve node execution
	nodeRequest := admin.NodeExecutionGetRequest{
		Id: &core.NodeExecutionIdentifier{
			NodeId:      nodeId,
			ExecutionId: executionId,
		},
	}

	nodeExecution, err := m.nodeExecutionManager.GetNodeExecution(ctx, nodeRequest)
	if err != nil {
		return nil, err
	}

	// populate cache and return
	nodeCache[nodeId] = nodeExecution
	return nodeExecution, nil
}

func (m *MetricsManager) parseExecution(ctx context.Context, execution *admin.Execution, depth int) (*admin.Span, error) {
	referenceSpan := &admin.ReferenceSpanInfo{
		Id: &admin.ReferenceSpanInfo_WorkflowId{
			WorkflowId: execution.Id,
		},
	}

	if depth != 0 {
		spans := make([]*admin.Span, 0) // TODO @hamersaw how to make an array
		nodeCache := make(map[string]*admin.NodeExecution)

		// retrieve workflow, execution, and node executions
		workflowRequest := admin.ObjectGetRequest{Id: execution.Closure.WorkflowId}
		workflow, err := m.workflowManager.GetWorkflow(ctx, workflowRequest)
		if err != nil {
			return nil, err
		}

		nodeExecutions := make([]*admin.NodeExecution, 0)
		nodeListRequest := admin.NodeExecutionListRequest{
			WorkflowExecutionId: execution.Id,
			Limit: 20, // TODO @hamersaw - parameterize?
		}

		for {
			nodeListResponse, err := m.nodeExecutionManager.ListNodeExecutions(ctx, nodeListRequest)
			if err != nil {
				return nil, err
			}

			for _, nodeExecution := range nodeListResponse.NodeExecutions {
				nodeExecutions = append(nodeExecutions, nodeExecution)
			}

			if len(nodeListResponse.NodeExecutions) < int(nodeListRequest.Limit) {
				break
			}

			nodeListRequest.Token = nodeListResponse.Token
		}

		// TODO @hamersaw - sort nodeExecutions by CreatedAt

		// compute frontend overhead
		startNode, err := m.getNodeExecution(ctx, "start-node", execution.Id, nodeCache)
		if err != nil {
			return nil, err
		}

		spans = append(spans, createCategoricalSpan(execution.Closure.CreatedAt,
			startNode.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))
		
		// iterate over nodes and compute overhead
		if err := m.parseNodeExecutions(ctx, nodeExecutions, &spans, depth); err != nil {
			return nil, err
		}

		// compute backend overhead
		latestUpstreamNode, err := m.getLatestUpstreamNodeExecution(ctx, "end-node", execution.Id,
			workflow.Closure.CompiledWorkflow.Primary.Connections.Upstream, nodeCache)
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

func (m *MetricsManager) parseNodeExecutions(ctx context.Context, nodeExecutions []*admin.NodeExecution, spans *[]*admin.Span, depth int) error {
	for _, nodeExecution := range nodeExecutions {
		if nodeExecution.Id.NodeId == "start-node" || nodeExecution.Id.NodeId == "end-node" {
			continue
		}

		nodeExecutionSpan, err := m.parseNodeExecution(ctx, nodeExecution, depth-1)
		if err != nil {
			return err
		}
		// TODO @hamersaw - prepend nodeExecution spans with NODE_TRANSITION time

		*spans = append(*spans, nodeExecutionSpan)
	}

	return nil
}

func (m *MetricsManager) parseNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution, depth int) (*admin.Span, error) {
	referenceSpan := &admin.ReferenceSpanInfo{
		Id: &admin.ReferenceSpanInfo_NodeId{
			NodeId: nodeExecution.Id,
		},
	}

	if depth != 0 {
		spans := make([]*admin.Span, 0) // TODO @hamersaw how to make an array

		taskExecutions := make([]*admin.TaskExecution, 0)
		taskListRequest := admin.TaskExecutionListRequest{
			NodeExecutionId: nodeExecution.Id,
			Limit: 20, // TODO @hamersaw - parameterize?
		}

		for {
			taskListResponse, err := m.taskExecutionManager.ListTaskExecutions(ctx, taskListRequest)
			if err != nil {
				return nil, err
			}

			for _, taskExecution := range taskListResponse.TaskExecutions {
				taskExecutions = append(taskExecutions, taskExecution)
			}

			if len(taskListResponse.TaskExecutions) < int(taskListRequest.Limit) {
				break
			}

			taskListRequest.Token = taskListResponse.Token
		}

		// TODO @hamersaw - sort taskExecutions by CreatedAt
		/*sort.Slice(a, func(i, j int) bool {
			return a[i] < a[j]
		})*/

		nodeExecutions := make([]*admin.NodeExecution, 0)
		nodeListRequest := admin.NodeExecutionListRequest{
			WorkflowExecutionId: nodeExecution.Id.ExecutionId,
			Limit: 20, // TODO @hamersaw - parameterize?
			UniqueParentId: nodeExecution.Id.NodeId,
		}

		// TODO - refactor this out!
		for {
			nodeListResponse, err := m.nodeExecutionManager.ListNodeExecutions(ctx, nodeListRequest)
			if err != nil {
				return nil, err
			}

			for _, nodeExecution := range nodeListResponse.NodeExecutions {
				nodeExecutions = append(nodeExecutions, nodeExecution)
			}

			if len(nodeListResponse.NodeExecutions) < int(nodeListRequest.Limit) {
				break
			}

			nodeListRequest.Token = nodeListResponse.Token
		}

		if !nodeExecution.Metadata.IsParentNode && len(taskExecutions) > 0 {
			// parse task node
			m.parseTaskNodeExecution(ctx, nodeExecution, taskExecutions, &spans, depth-1)
		} else if nodeExecution.Metadata.IsParentNode && len(taskExecutions) > 0 {
			// TODO @hamersaw - dynamic node
			if err := m.parseDynamicNodeExecution(ctx, nodeExecution, taskExecutions, nodeExecutions, &spans, depth-1); err != nil {
				return nil, err
			}
		//} else if nodeExecution.Metadata.IsParentNode && nodeExecution.Closure.Target? is a WorkflowNode
		//} else if nodeExecution.Metadata.IsParentNode && HAS_UNDERLYING_NODE_EXECUTIONS
		} else {
			// TODO @hamersaw process branch, gate, launchplan, subworkflow
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

func (m *MetricsManager) parseDynamicNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution,
	taskExecutions []*admin.TaskExecution, nodeExecutions []*admin.NodeExecution, spans *[]*admin.Span, depth int) error {

	*spans = append(*spans, createCategoricalSpan(nodeExecution.Closure.CreatedAt,
		taskExecutions[0].Closure.CreatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	parseTaskExecutions(taskExecutions, spans, depth)

	// TODO @hamersaw - frontend overhead
	//*spans = append(*spans, createCategoricalSpan(taskExecutions[len(taskExecutions)-1].Closure.UpdatedAt,
	//	nodeExecution.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	// TODO @hamersaw - correctly parse dynamic nodes
	if err := m.parseNodeExecutions(ctx, nodeExecutions, spans, depth); err != nil {
		return err
	}

	// TODO @hamersaw - backend overhead

	return nil
}

func (m *MetricsManager) parseTaskNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution,
	taskExecutions []*admin.TaskExecution, spans *[]*admin.Span, depth int) {

	*spans = append(*spans, createCategoricalSpan(nodeExecution.Closure.CreatedAt,
		taskExecutions[0].Closure.CreatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))

	parseTaskExecutions(taskExecutions, spans, depth)

	*spans = append(*spans, createCategoricalSpan(taskExecutions[len(taskExecutions)-1].Closure.UpdatedAt,
		nodeExecution.Closure.UpdatedAt, admin.CategoricalSpanInfo_EXECUTION_OVERHEAD))
}

func parseTaskExecutions(taskExecutions []*admin.TaskExecution, spans *[]*admin.Span, depth int) {
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

	// retrieve node and task executions
	nodeCache := make(map[string]*admin.NodeExecution)
	nodeExecution, err := m.getNodeExecution(ctx, request.Id.NodeId, request.Id.ExecutionId, nodeCache)
	if err != nil {
		return nil, err
	}

	span, err := m.parseNodeExecution(ctx, nodeExecution, int(request.Depth))
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
