package models

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

// Task primary key
type TaskKey struct {
	Project string `gorm:"primary_key;index:task_project_domain_name_idx,task_project_domain_idx"`
	Domain  string `gorm:"primary_key;index:task_project_domain_name_idx,task_project_domain_idx"`
	Name    string `gorm:"primary_key;index:task_project_domain_name_idx"`
	Version string `gorm:"primary_key"`
}

// Database model to encapsulate a task.
type Task struct {
	BaseModel
	TaskKey
	Closure []byte `gorm:"not null"`
	// Hash of the compiled task closure
	Digest []byte
	// Task type (also stored in the closure put promoted as a column for filtering).
	Type string
}
