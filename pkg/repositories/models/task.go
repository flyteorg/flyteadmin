package models

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

// Task primary key
type TaskKey struct {
	Project string `gorm:"primary_key;index:task_project_domain_name_idx;index:task_project_domain_idx" valid:"length(0|255)"`
	Domain  string `gorm:"primary_key;index:task_project_domain_name_idx;index:task_project_domain_idx" valid:"length(0|255)"`
	Name    string `gorm:"primary_key;index:task_project_domain_name_idx" valid:"length(0|255)"`
	Version string `gorm:"primary_key" valid:"length(0|255)"`
}

// Database model to encapsulate a task.
type Task struct {
	BaseModel
	TaskKey
	Closure []byte `gorm:"not null"`
	// Hash of the compiled task closure
	Digest []byte
	// Task type (also stored in the closure put promoted as a column for filtering).
	Type          string `valid:"length(0|255)"`
	DescriptionID uint   `gorm:"index"`
	// ShortDescription is saved in the description entity table. set this to read only so we won't create this column.
	// Adding ShortDescription because we want to unmarshal the short description in the
	// descriptionEntity table to workflow object.
	ShortDescription string `gorm:"<-:false"`
}
