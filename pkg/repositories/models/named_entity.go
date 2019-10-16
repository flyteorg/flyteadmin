package models

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

// IMPORTANT: If you update the model below, be sure to double check model definitions in
// pkg/repositories/config/migration_models.go

// NamedEntityMetadata primary key
type NamedEntityMetadataKey struct {
	ResourceType core.ResourceType `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
	Project      string            `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
	Domain       string            `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
	Name         string            `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
}

// Fields to be composed into any named entity
type NamedEntityMetadataFields struct {
	Description string `gorm:"type:varchar(300)"`
}

// Database model to encapsulate metadata associated with a NamedEntity
type NamedEntityMetadata struct {
	BaseModel
	NamedEntityMetadataKey
	NamedEntityMetadataFields
}

type NamedEntityKey struct {
	ResourceType core.ResourceType
	Project      string
	Domain       string
	Name         string
}

type NamedEntity struct {
	NamedEntityKey
	NamedEntityMetadataFields
}
