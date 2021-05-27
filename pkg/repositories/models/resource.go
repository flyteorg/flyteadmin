package models

import "time"

type ResourcePriority int32

const (
	ResourcePriorityDomainLevel        ResourcePriority = 1
	ResourcePriorityProjectDomainLevel ResourcePriority = 10
	ResourcePriorityWorkflowLevel      ResourcePriority = 100
	ResourcePriorityLaunchPlanLevel    ResourcePriority = 1000
)

// Represents Flyte resources repository.
// In this model, the combination of (Project, Domain, Workflow, LaunchPlan, ResourceType) is unique
type Resource struct {
	ID           int64 `gorm:"AUTO_INCREMENT;column:id;primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `sql:"index"`
	Project      string     `gorm:"unique_index:resource_idx" valid:"length(1|50)"`
	Domain       string     `gorm:"unique_index:resource_idx" valid:"length(1|50)"`
	Workflow     string     `gorm:"unique_index:resource_idx" valid:"length(1|50)"`
	LaunchPlan   string     `gorm:"unique_index:resource_idx" valid:"length(1|50)"`
	ResourceType string     `gorm:"unique_index:resource_idx" valid:"length(1|50)"`
	Priority     ResourcePriority
	// Serialized flyteidl.admin.MatchingAttributes.
	Attributes []byte
}
