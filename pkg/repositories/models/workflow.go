package models

// Workflow primary key
// TODO: Add size to the fields
type WorkflowKey struct {
	Project string `gorm:"size:64;primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Domain  string `gorm:"size:255;primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Name    string `gorm:"size:255;primary_key;index:workflow_project_domain_name_idx"  valid:"length(0|255)"`
	Version string `gorm:"size:255;primary_key"`
}

// Database model to encapsulate a workflow.
type Workflow struct {
	BaseModel
	WorkflowKey
	TypedInterface          []byte
	RemoteClosureIdentifier string `gorm:"size:255;not null"`
	// Hash of the compiled workflow closure
	Digest []byte
	// ShortDescription for the workflow.
	ShortDescription string	`gorm:"size:255"`
}
