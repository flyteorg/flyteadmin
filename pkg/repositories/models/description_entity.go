package models

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

// DescriptionEntityKey DescriptionEntity primary key
type DescriptionEntityKey struct {
	ResourceType core.ResourceType `gorm:"primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
	Project      string            `gorm:"size:100;primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
	Domain       string            `gorm:"size:100;primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
	Name         string            `gorm:"size:100;primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
	Version      string            `gorm:"size:100;primary_key;index:description_entity_project_domain_name_version_idx" valid:"length(0|255)"`
}

// SourceCode Database model to encapsulate a SourceCode.
type SourceCode struct {
	Link string `gorm:"size:100" valid:"length(0|255)"`
}

// DescriptionEntity Database model to encapsulate a DescriptionEntity.
type DescriptionEntity struct {
	DescriptionEntityKey

	BaseModel

	SourceCode

	ShortDescription string `gorm:"size:100`

	LongDescription []byte
}
