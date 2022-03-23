package config

import (
	"database/sql"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	schedulerModels "github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// TODO: add a way to get these list of tables directly from the gorm loaded models
var (
	tables = []string{"execution_events", "executions", "launch_plans", "named_entity_metadata",
		"node_execution_events", "node_executions", "projects", "resources", "schedulable_entities",
		"schedule_entities_snapshots", "task_executions", "tasks", "workflows"}
)

type MigrationOptions struct {
	IgnoreForSqlite bool
}

type Migration struct {
	gormigrate.Migration
	Options MigrationOptions
}

var Migrations = []*Migration{
	// Create projects table.
	{
		Migration: gormigrate.Migration{
			ID: "2019-05-22-projects",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Project{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("projects")
			},
		},
	},
	// Create Task
	{
		Migration: gormigrate.Migration{
			ID: "2018-05-23-tasks",
			Migrate: func(tx *gorm.DB) error {
				// The gormigrate library recommends that we copy the actual struct into here for record-keeping but after
				// some internal discussion we've decided that that's not necessary. Just a history of what we've touched
				// when should be sufficient.
				return tx.AutoMigrate(&models.Task{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("tasks")
			},
		},
	},
	// Create Workflow
	{
		Migration: gormigrate.Migration{
			ID: "2018-05-23-workflows",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Workflow{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("workflows")
			},
		},
	},
	// Create Launch Plan table
	{
		Migration: gormigrate.Migration{
			ID: "2019-05-23-lp",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.LaunchPlan{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("launch_plans")
			},
		},
	},

	// Create executions table
	{
		Migration: gormigrate.Migration{
			ID: "2019-05-23-executions",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Execution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("executions")
			},
		},
	},
	// Create executions events table
	{
		Migration: gormigrate.Migration{
			ID: "2019-01-29-executions-events",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&ExecutionEvent{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("executions_events")
			},
		},
	},

	// Create node executions table
	{
		Migration: gormigrate.Migration{
			ID: "2019-04-17-node-executions",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&NodeExecution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("node_executions")
			},
		},
	},
	// Create node executions events table
	{
		Migration: gormigrate.Migration{
			ID: "2019-01-29-node-executions-events",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&NodeExecutionEvent{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("node_executions_events")
			},
		},
	},
	// Create task executions table
	{
		Migration: gormigrate.Migration{
			ID: "2019-03-16-task-executions",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&TaskExecution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("task_executions")
			},
		},
	},
	// Update node executions with null parent values
	{
		Migration: gormigrate.Migration{
			ID: "2019-04-17-node-executions-backfill",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("update node_executions set parent_task_execution_id = NULL where parent_task_execution_id = 0").Error
			},
		},
	},
	// Update executions table to add cluster
	{
		Migration: gormigrate.Migration{
			ID: "2019-09-27-executions",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Execution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS cluster").Error
			},
		},
	},
	// Update projects table to add description column
	{
		Migration: gormigrate.Migration{
			ID: "2019-10-09-project-description",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Project{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE projects DROP COLUMN IF EXISTS description").Error
			},
		},
	},
	// Add offloaded URIs to table
	{
		Migration: gormigrate.Migration{
			ID: "2019-10-15-offload-inputs",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Execution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS InputsURI, DROP COLUMN IF EXISTS UserInputsURI").Error
			},
		},
	},
	// Create named_entity_metadata table.
	{
		Migration: gormigrate.Migration{
			ID: "2019-11-05-named-entity-metadata",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.NamedEntityMetadata{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("named_entity_metadata")
			},
		},
	},
	// Add ProjectAttributes with custom resource attributes.
	{
		Migration: gormigrate.Migration{
			ID: "2020-01-10-resource",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Resource{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("resources")
			},
		},
	},
	// Add Type to Task model.
	{
		Migration: gormigrate.Migration{
			ID: "2020-03-17-task-type",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Task{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE tasks DROP COLUMN IF EXISTS type").Error
			},
		},
	},
	// Add state to name entity model
	{
		Migration: gormigrate.Migration{
			ID: "2020-04-03-named-entity-state",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.NamedEntityMetadata{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Table("named_entity_metadata").Migrator().DropColumn(&models.NamedEntityMetadata{}, "state")
			},
		},
	},
	// Set default state value for workflow model
	{
		Migration: gormigrate.Migration{
			ID: "2020-04-03-named-entity-state-default",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("UPDATE named_entity_metadata SET state = 0").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("UPDATE named_entity_metadata set state = NULL").Error
			},
		},
	},
	// Modify the workflows table, if necessary
	{
		Migration: gormigrate.Migration{
			ID: "2020-04-03-workflow-state",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE workflows DROP COLUMN IF EXISTS state").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE workflows ADD COLUMN IF NOT EXISTS state integer;").Error
			},
		},
		Options: MigrationOptions{IgnoreForSqlite: true},
	},
	// Modify the executions & node_execution table, if necessary
	{
		Migration: gormigrate.Migration{
			ID: "2020-04-29-executions",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Execution{}, &NodeExecution{})
			},
			Rollback: func(tx *gorm.DB) error {
				if err := tx.Model(&models.Execution{}).Migrator().DropColumn(&models.Execution{}, "error_code"); err != nil {
					return err
				}
				if err := tx.Model(&models.Execution{}).Migrator().DropColumn(&models.Execution{}, "error_kind"); err != nil {
					return err
				}
				if err := tx.Model(&NodeExecution{}).Migrator().DropColumn(&NodeExecution{}, "error_code"); err != nil {
					return err
				}
				if err := tx.Model(&NodeExecution{}).Migrator().DropColumn(&NodeExecution{}, "error_kind"); err != nil {
					return err
				}
				return nil
			},
		},
	},
	// Add TaskID to Execution model.
	{
		Migration: gormigrate.Migration{
			ID: "2020-04-14-task-type",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Execution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE executions DROP COLUMN IF EXISTS task_id").Error
			},
		},
	},

	// NodeExecutions table has CacheStatus for Task nodes
	{
		Migration: gormigrate.Migration{
			ID: "2020-07-27-cachestatus",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&NodeExecution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&NodeExecution{}).Migrator().DropColumn(&NodeExecution{}, "cache_status")
			},
		},
	},
	{
		Migration: gormigrate.Migration{
			ID: "2020-07-31-node-execution",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&NodeExecution{})
			},
			Rollback: func(tx *gorm.DB) error {
				if err := tx.Model(&NodeExecution{}).Migrator().DropColumn(&NodeExecution{}, "parent_id"); err != nil {
					return err
				}
				return tx.Model(&NodeExecution{}).Migrator().DropColumn(&NodeExecution{}, "node_execution_metadata")
			},
		},
	},
	{
		Migration: gormigrate.Migration{
			ID: "2020-08-17-labels-addition",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Project{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&models.Project{}).Migrator().DropColumn(&models.Project{}, "labels")
			},
		},
	},
	{
		Migration: gormigrate.Migration{
			ID: "2020-09-01-task-exec-idx",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&TaskExecution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&TaskExecution{}).Migrator().DropIndex(&TaskExecution{}, "idx_task_executions_exec")
			},
		},
	},
	{
		Migration: gormigrate.Migration{
			ID: "2020-11-03-project-state-addition",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Project{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&models.Project{}).Migrator().DropColumn(&models.Project{}, "state")
			},
		},
	},
	{
		Migration: gormigrate.Migration{
			ID: "2020-11-03-project-state-default",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("UPDATE projects set state = 0").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("UPDATE projects set state = NULL").Error
			},
		},
	},
	{
		Migration: gormigrate.Migration{
			ID: "2021-01-22-execution-user",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Execution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&models.Execution{}).Migrator().DropColumn(&models.Execution{}, "user")
			},
		},
	},
	{
		Migration: gormigrate.Migration{
			ID: "2021-04-19-node-execution_dynamic-workflow",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&NodeExecution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Model(&NodeExecution{}).Migrator().DropColumn(&NodeExecution{}, "dynamic_workflow_remote_closure_reference")
			},
		},
	},

	{
		Migration: gormigrate.Migration{
			ID: "2021-07-22-schedulable_entities",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&schedulerModels.SchedulableEntity{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(&schedulerModels.SchedulableEntity{}, "schedulable_entities")
			},
		},
	},

	{
		Migration: gormigrate.Migration{
			ID: "2021-08-05-schedulable_entities_snapshot",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&schedulerModels.ScheduleEntitiesSnapshot{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(&schedulerModels.ScheduleEntitiesSnapshot{}, "schedulable_entities_snapshot")
			},
		},
	},

	// For any new table, Please use the following pattern due to a bug
	// in the postgres gorm layer https://github.com/go-gorm/postgres/issues/65
	{
		Migration: gormigrate.Migration{
			ID: "2022-01-11-id-to-bigint",
			Migrate: func(tx *gorm.DB) error {
				db, err := tx.DB()
				if err != nil {
					return err
				}
				return alterTableColumnType(db, "id", "bigint")
			},
			Rollback: func(tx *gorm.DB) error {
				db, err := tx.DB()
				if err != nil {
					return err
				}
				return alterTableColumnType(db, "id", "int")
			},
		},
		Options: MigrationOptions{IgnoreForSqlite: true},
	},

	// Add state to execution model.
	{
		Migration: gormigrate.Migration{
			ID: "2022-01-11-execution-state",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Execution{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Table("execution").Migrator().DropColumn(&models.Execution{}, "state")
			},
		},
	},
}

func alterTableColumnType(db *sql.DB, columnName, columnType string) error {

	var err error
	for _, table := range tables {
		if _, err = db.Exec(fmt.Sprintf(`ALTER TABLE IF EXISTS %s ALTER COLUMN "%s" TYPE %s`, table, columnName,
			columnType)); err != nil {
			return err
		}
	}
	return nil
}
