package models

import (
	"time"
)

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

// Execution primary key
type ExecutionKey struct {
	Project string `gorm:"primary_key;column:execution_project"`
	Domain  string `gorm:"primary_key;column:execution_domain"`
	Name    string `gorm:"primary_key;column:execution_name"`
}

// Database model to encapsulate a (workflow) execution.
type Execution struct {
	BaseModel
	ExecutionKey
	LaunchPlanID uint `gorm:"index"`
	WorkflowID   uint `gorm:"index"`
	Phase        string
	Closure      []byte
	Spec         []byte `gorm:"not null"`
	StartedAt    *time.Time
	// Corresponds to the CreatedAt field in the Execution closure.
	// Prefixed with Execution to avoid clashes with gorm.Model CreatedAt
	ExecutionCreatedAt *time.Time
	// Corresponds to the UpdatedAt field in the Execution closure
	// Prefixed with Execution to avoid clashes with gorm.Model UpdatedAt
	ExecutionUpdatedAt *time.Time
	Duration           time.Duration
	ExecutionEvents    []ExecutionEvent
	// In the case of an aborted execution this string may be non-empty.
	// It should be ignored for any other value of phase other than aborted.
	AbortCause string
	// Corresponds to the execution mode used to trigger this execution
	Mode int32
	// The "parent" execution (if there is one) that is related to this execution.
	SourceExecutionID uint
	// Descendant execution related to this execution.
	DescendantExecution *Execution
	// The parent node execution if this was launched by a node
	ParentNodeExecutionID uint
	// Cluster where execution was triggered
	Cluster string
}
