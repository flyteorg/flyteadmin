package models

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

// DescriptionEntityKey DescriptionEntity primary key
type DescriptionEntityKey struct {
	ResourceType core.ResourceType `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
	Project      string            `gorm:"primary_key;index:task_project_domain_name_idx;index:task_project_domain_idx" valid:"length(0|255)"`
	Domain       string            `gorm:"primary_key;index:task_project_domain_name_idx;index:task_project_domain_idx" valid:"length(0|255)"`
	Name         string            `gorm:"primary_key;index:task_project_domain_name_idx" valid:"length(0|255)"`
	Version      string            `gorm:"primary_key" valid:"length(0|255)"`
}

// SourceCode Database model to encapsulate a SourceCode.
type SourceCode struct {
	Link string `valid:"length(0|255)"`
}

// DescriptionEntity Database model to encapsulate a DescriptionEntity.
type DescriptionEntity struct {
	DescriptionEntityKey

	BaseModel
	// Hash of the Description entity
	Digest []byte

	ShortDescription string `valid:"length(0|255)"`

	Tags string `valid:"length(0|255)"`

	Labels string `valid:"length(0|255)"`

	SourceCode
}
