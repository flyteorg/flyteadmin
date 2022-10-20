package models

// Workflow primary key
type WorkflowKey struct {
	Project string `gorm:"primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Domain  string `gorm:"primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Name    string `gorm:"primary_key;index:workflow_project_domain_name_idx"  valid:"length(0|255)"`
	Version string `gorm:"primary_key"`
}

// Database model to encapsulate a workflow.
type Workflow struct {
	BaseModel
	WorkflowKey
	TypedInterface          []byte
	RemoteClosureIdentifier string `gorm:"not null" valid:"length(0|255)"`
	// Hash of the compiled workflow closure
	Digest        []byte
	DescriptionID uint `gorm:"index"`
	// ShortDescription is saved in the description entity table. set this to read only so we won't create this column.
	// Adding ShortDescription because we want to unmarshal the short description in the
	// descriptionEntity table to workflow object.
	ShortDescription string `gorm:"<-:false"`
}
