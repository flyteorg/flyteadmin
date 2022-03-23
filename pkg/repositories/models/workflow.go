package models

// Workflow primary key
type WorkflowKey struct {
	Project string `gorm:"uniqueIndex:primary_workflow_index;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Domain  string `gorm:"uniqueIndex:primary_workflow_index;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Name    string `gorm:"uniqueIndex:primary_workflow_index;index:workflow_project_domain_name_idx"  valid:"length(0|255)"`
	Version string `gorm:"uniqueIndex:primary_workflow_index"`
}

// Database model to encapsulate a workflow.
type Workflow struct {
	BaseModel
	WorkflowKey
	TypedInterface          []byte
	RemoteClosureIdentifier string `gorm:"not null" valid:"length(0|255)"`
	// Hash of the compiled workflow closure
	Digest []byte
}
