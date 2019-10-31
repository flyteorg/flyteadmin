package gormimpl

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/lyft/flyteadmin/pkg/common"
	adminErrors "github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
)

const Project = "project"
const Domain = "domain"
const Name = "name"
const Version = "version"
const Closure = "closure"
const Description = "description"

const ProjectID = "project_id"
const ProjectName = "project_name"
const DomainID = "domain_id"
const DomainName = "domain_name"

const executionTableName = "executions"
const namedEntityMetadataTableName = "named_entity_metadata"
const nodeExecutionTableName = "node_executions"
const nodeExecutionEventTableName = "node_event_executions"
const taskExecutionTableName = "task_executions"
const taskTableName = "tasks"

const limit = "limit"
const filters = "filters"

var identifierGroupBy = fmt.Sprintf("%s, %s, %s", Project, Domain, Name)

var innerJoinNodeExecToNodeEvents = fmt.Sprintf(
	"INNER JOIN %s ON %s.node_execution_id = %s.id",
	nodeExecutionTableName, nodeExecutionEventTableName, nodeExecutionTableName)

var innerJoinExecToNodeExec = fmt.Sprintf(
	"INNER JOIN %s ON %s.execution_project = %s.execution_project AND "+
		"%s.execution_domain = %s.execution_domain AND %s.execution_name = %s.execution_name",
	executionTableName, nodeExecutionTableName, executionTableName, nodeExecutionTableName, executionTableName,
	nodeExecutionTableName, executionTableName)

var innerJoinNodeExecToTaskExec = fmt.Sprintf(
	"INNER JOIN %s ON %s.node_id = %s.node_id AND %s.execution_project = %s.execution_project AND "+
		"%s.execution_domain = %s.execution_domain AND %s.execution_name = %s.execution_name",
	nodeExecutionTableName, taskExecutionTableName, nodeExecutionTableName, taskExecutionTableName,
	nodeExecutionTableName, taskExecutionTableName, nodeExecutionTableName, taskExecutionTableName,
	nodeExecutionTableName)

// Because dynamic tasks do NOT necessarily register static task definitions, we use a left join to not exclude
// dynamic tasks from list queries.
var leftJoinTaskToTaskExec = fmt.Sprintf(
	"LEFT JOIN %s ON %s.project = %s.project AND %s.domain = %s.domain AND %s.name = %s.name AND "+
		"%s.version = %s.version",
	taskTableName, taskExecutionTableName, taskTableName, taskExecutionTableName, taskTableName,
	taskExecutionTableName, taskTableName, taskExecutionTableName, taskTableName)

// Creates a JOIN clause between a table which can hold NamedEntityIdentifiers
// and the metadata associated between them. Metadata is optional, so a left
// join is used.
func CreateEntityMetadataJoin(entityTableName string, resourceType core.ResourceType) string {
	return fmt.Sprintf(
		"LEFT JOIN %s ON %s.resource_type = %d AND %s.project = %s.project AND %s.domain = %s.domain AND %s.name = %s.name", namedEntityMetadataTableName, namedEntityMetadataTableName, resourceType, namedEntityMetadataTableName, entityTableName,
		namedEntityMetadataTableName, entityTableName,
		namedEntityMetadataTableName, entityTableName)
}

var leftJoinWorkflowNameToMetadata = CreateEntityMetadataJoin(workflowTableName, core.ResourceType_WORKFLOW)
var leftJoinLaunchPlanNameToMetadata = CreateEntityMetadataJoin(launchPlanTableName, core.ResourceType_LAUNCH_PLAN)
var leftJoinTaskNameToMetadata = CreateEntityMetadataJoin(taskTableName, core.ResourceType_TASK)

var resourceTypeToTableName = map[core.ResourceType]string{
	core.ResourceType_LAUNCH_PLAN: launchPlanTableName,
	core.ResourceType_WORKFLOW:    workflowTableName,
	core.ResourceType_TASK:        taskTableName,
}

var resourceTypeToMetadataJoin = map[core.ResourceType]string{
	core.ResourceType_LAUNCH_PLAN: leftJoinLaunchPlanNameToMetadata,
	core.ResourceType_WORKFLOW:    leftJoinWorkflowNameToMetadata,
	core.ResourceType_TASK:        leftJoinTaskNameToMetadata,
}

var entityToModel = map[common.Entity]interface{}{
	common.Execution:          models.Execution{},
	common.LaunchPlan:         models.LaunchPlan{},
	common.NodeExecution:      models.NodeExecution{},
	common.NodeExecutionEvent: models.NodeExecutionEvent{},
	common.Task:               models.Task{},
	common.TaskExecution:      models.TaskExecution{},
	common.Workflow:           models.Workflow{},
}

// Validates there are no missing but required parameters in ListResourceInput
func ValidateListInput(input interfaces.ListResourceInput) adminErrors.FlyteAdminError {
	if input.Limit == 0 {
		return errors.GetInvalidInputError(limit)
	}
	if len(input.InlineFilters) == 0 {
		return errors.GetInvalidInputError(filters)
	}
	return nil
}

func applyFilters(tx *gorm.DB, inlineFilters []common.InlineFilter, mapFilters []common.MapFilter) (*gorm.DB, error) {
	for _, filter := range inlineFilters {
		gormQueryExpr, err := filter.GetGormQueryExpr()
		if err != nil {
			return nil, errors.GetInvalidInputError(err.Error())
		}
		tx = tx.Where(gormQueryExpr.Query, gormQueryExpr.Args)
	}
	for _, mapFilter := range mapFilters {
		tx = tx.Where(mapFilter.GetFilter())
	}
	return tx, nil
}

func applyScopedFilters(tx *gorm.DB, inlineFilters []common.InlineFilter, mapFilters []common.MapFilter) (*gorm.DB, error) {
	for _, filter := range inlineFilters {
		entityModel, ok := entityToModel[filter.GetEntity()]
		if !ok {
			return nil, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"unrecognized entity in filter expression: %v", filter.GetEntity())
		}
		tableName := tx.NewScope(entityModel).TableName()
		gormQueryExpr, err := filter.GetGormJoinTableQueryExpr(tableName)
		if err != nil {
			return nil, err
		}
		tx = tx.Where(gormQueryExpr.Query, gormQueryExpr.Args)
	}
	for _, mapFilter := range mapFilters {
		tx = tx.Where(mapFilter.GetFilter())
	}
	return tx, nil
}

func getGroupByForNamedEntity(tableName string) string {
	return fmt.Sprintf("%s.%s, %s.%s, %s.%s, %s.%s", tableName, Project, tableName, Domain, tableName, Name, namedEntityMetadataTableName, Description)
}

func getSelectForNamedEntity(tableName string) []string {
	return []string{
		fmt.Sprintf("%s.%s", tableName, Project),
		fmt.Sprintf("%s.%s", tableName, Domain),
		fmt.Sprintf("%s.%s", tableName, Name),
		fmt.Sprintf("%s.%s", namedEntityMetadataTableName, Description),
	}
}

func getNamedEntityFilters(resourceType core.ResourceType, project string, domain string, name string) ([]common.InlineFilter, error) {
	entity := common.ResourceTypeToEntity[resourceType]

	filters := make([]common.InlineFilter, 0)
	projectFilter, err := common.NewSingleValueFilter(entity, common.Equal, Project, project)
	if err != nil {
		return nil, err
	}
	filters = append(filters, projectFilter)
	domainFilter, err := common.NewSingleValueFilter(entity, common.Equal, Domain, domain)
	if err != nil {
		return nil, err
	}
	filters = append(filters, domainFilter)
	nameFilter, err := common.NewSingleValueFilter(entity, common.Equal, Name, name)
	if err != nil {
		return nil, err
	}
	filters = append(filters, nameFilter)

	return filters, nil
}
