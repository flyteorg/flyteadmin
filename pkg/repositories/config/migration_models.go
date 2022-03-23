package config

import (
	"time"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

/*
	IMPORTANT: You'll observe several models are redefined below with named index tags *omitted*. This is because
	postgres requires that index names be unique across *all* tables. If you modify Task, Execution, NodeExecution or
	TaskExecution models in code be sure to update the appropriate duplicate definitions here.
*/

type TaskKey struct {
	Project string `gorm:"uniqueIndex:primary_task_exec_index"`
	Domain  string `gorm:"uniqueIndex:primary_task_exec_index"`
	Name    string `gorm:"uniqueIndex:primary_task_exec_index"`
	Version string `gorm:"uniqueIndex:primary_task_exec_index"`
}

type NodeExecutionKey struct {
	Project string `gorm:"uniqueIndex:primary_node_exec_index;column:execution_project"`
	Domain  string `gorm:"uniqueIndex:primary_node_exec_index;column:execution_domain"`
	Name    string `gorm:"uniqueIndex:primary_node_exec_index;column:execution_name"`
	NodeID  string `gorm:"uniqueIndex:primary_node_exec_index;index"`
}

type NodeExecution struct {
	models.BaseModel
	NodeExecutionKey
	// Also stored in the closure, but defined as a separate column because it's useful for filtering and sorting.
	Phase     string
	InputURI  string
	Closure   []byte
	StartedAt *time.Time
	// Corresponds to the CreatedAt field in the NodeExecution closure
	// Prefixed with NodeExecution to avoid clashes with gorm.Model CreatedAt
	NodeExecutionCreatedAt *time.Time
	// Corresponds to the UpdatedAt field in the NodeExecution closure
	// Prefixed with NodeExecution to avoid clashes with gorm.Model UpdatedAt
	NodeExecutionUpdatedAt *time.Time
	Duration               time.Duration
	// Metadata about the node execution.
	NodeExecutionMetadata []byte
	// Parent that spawned this node execution - value is empty for executions at level 0
	ParentID *uint `sql:"default:null" gorm:"index"`
	// List of child node executions - for cases like Dynamic task, sub workflow, etc
	ChildNodeExecutions []NodeExecution `gorm:"foreignKey:ParentID;references:ID"`
	// The task execution (if any) which launched this node execution.
	// TO BE DEPRECATED - as we have now introduced ParentID
	ParentTaskExecutionID *uint `sql:"default:null" gorm:"index"`
	// The workflow execution (if any) which this node execution launched
	// NOTE: LaunchedExecution[foreignkey:ParentNodeExecutionID] refers to Workflow execution launched and is different from ParentID
	LaunchedExecution models.Execution `gorm:"foreignKey:ParentNodeExecutionID;references:ID"`
	// Execution Error Kind. nullable, can be one of core.ExecutionError_ErrorKind
	ErrorKind *string `gorm:"index"`
	// Execution Error Code nullable. string value, but finite set determined by the execution engine and plugins
	ErrorCode *string
	// If the node is of Type Task, this should always exist for a successful execution, indicating the cache status for the execution
	CacheStatus *string
	// In the case of dynamic workflow nodes, the remote closure is uploaded to the path specified here.
	DynamicWorkflowRemoteClosureReference string
}

type TaskExecutionKey struct {
	TaskKey
	Project string `gorm:"uniqueIndex:primary_task_exec_index;column:execution_project;index:idx_task_executions_exec"`
	Domain  string `gorm:"uniqueIndex:primary_task_exec_index;column:execution_domain;index:idx_task_executions_exec"`
	Name    string `gorm:"uniqueIndex:primary_task_exec_index;column:execution_name;index:idx_task_executions_exec"`
	NodeID  string `gorm:"uniqueIndex:primary_task_exec_index;index:idx_task_executions_exec;index"`
	// *IMPORTANT* This is a pointer to an int in order to allow setting an empty ("0") value according to gorm convention.
	// Because RetryAttempt is part of the TaskExecution primary key is should *never* be null.
	RetryAttempt *uint32 `gorm:"uniqueIndex:primary_task_exec_index;AUTO_INCREMENT:FALSE"`
}

type TaskExecution struct {
	models.BaseModel
	TaskExecutionKey
	Phase        string
	PhaseVersion uint32
	InputURI     string
	Closure      []byte
	StartedAt    *time.Time
	// Corresponds to the CreatedAt field in the TaskExecution closure
	// This field is prefixed with TaskExecution because it signifies when
	// the execution was createdAt, not to be confused with gorm.Model.CreatedAt
	TaskExecutionCreatedAt *time.Time
	// Corresponds to the UpdatedAt field in the TaskExecution closure
	// This field is prefixed with TaskExecution because it signifies when
	// the execution was UpdatedAt, not to be confused with gorm.Model.UpdatedAt
	TaskExecutionUpdatedAt *time.Time
	Duration               time.Duration
	// The child node executions (if any) launched by this task execution.
	ChildNodeExecution []NodeExecution `gorm:"foreignkey:ParentTaskExecutionID;references:ID"`
}

type ExecutionEvent struct {
	models.BaseModel
	Project    string `gorm:"uniqueIndex:primary_ee_index;column:execution_project"`
	Domain     string `gorm:"uniqueIndex:primary_ee_index;column:execution_domain"`
	Name       string `gorm:"uniqueIndex:primary_ee_index;column:execution_name"`
	RequestID  string `valid:"length(0|255)"`
	OccurredAt time.Time
	Phase      string `gorm:"uniqueIndex:primary_ee_index"`
}

type NodeExecutionEvent struct {
	models.BaseModel
	Project    string `gorm:"uniqueIndex:primary_nee_index;column:execution_project"`
	Domain     string `gorm:"uniqueIndex:primary_nee_index;column:execution_domain"`
	Name       string `gorm:"uniqueIndex:primary_nee_index;column:execution_name"`
	NodeID     string `gorm:"uniqueIndex:primary_nee_index;index"`
	RequestID  string
	OccurredAt time.Time
	Phase      string `gorm:"uniqueIndex:primary_nee_index"`
}
