package config

import (
	"github.com/jinzhu/gorm"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	gormigrate "gopkg.in/gormigrate.v1"
)

var Migrations = []*gormigrate.Migration{
	// Create projects table.
	{
		ID: "2019-05-22-projects",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Project{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("projects").Error
		},
	},
	// Create Task
	{
		ID: "2018-05-23-tasks",
		Migrate: func(tx *gorm.DB) error {
			// The gormigrate library recommends that we copy the actual struct into here for record-keeping but after
			// some internal discussion we've decided that that's not necessary. Just a history of what we've touched
			// when should be sufficient.
			return tx.AutoMigrate(&models.Task{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("tasks").Error
		},
	},
	// Create Workflow
	{
		ID: "2018-05-23-workflows",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Workflow{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("workflows").Error
		},
	},
	// Create Launch Plan table
	{
		ID: "2019-05-23-lp",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.LaunchPlan{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("launch_plans").Error
		},
	},

	// Create executions table
	{
		ID: "2019-05-23-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("executions").Error
		},
	},
	// Create executions events table
	{
		ID: "2019-01-29-executions-events",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.ExecutionEvent{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("executions_events").Error
		},
	},

	// Create node executions table
	{
		ID: "2019-04-17-node-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&NodeExecution{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("node_executions").Error
		},
	},
	// Create node executions events table
	{
		ID: "2019-01-29-node-executions-events",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NodeExecutionEvent{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("node_executions_events").Error
		},
	},
	// Create task executions table
	{
		ID: "2019-03-16-task-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&TaskExecution{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("task_executions").Error
		},
	},
	// Update node executions with null parent values
	{
		ID: "2019-04-17-node-executions-backfill",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("update node_executions set parent_task_execution_id = NULL where parent_task_execution_id = 0").Error
		},
	},
	// Update executions table to add cluster
	{
		ID: "2019-09-27-executions",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS cluster").Error
		},
	},
	// Update projects table to add description column
	{
		ID: "2019-10-09-project-description",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Project{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE projects DROP COLUMN IF EXISTS description").Error
		},
	},
	// Add offloaded URIs to table
	{
		ID: "2019-10-15-offload-inputs",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Execution{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS InputsURI, DROP COLUMN IF EXISTS UserInputsURI").Error
		},
	},
	// Add ProjectDomains with custom resource attributes.
	{
		ID: "2019-10-28-project-domains",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.ProjectDomain{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("project_domains").Error
		},
	},
	// Create named_entity_metadata table.
	{
		ID: "2019-11-05-named-entity-metadata",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.NamedEntityMetadata{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("named_entity_metadata").Error
		},
	},
}
