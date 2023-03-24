package models

import "time"

type ResourcePriority int32

const (
	ResourcePriorityProjectLevel       ResourcePriority = 5 // use this
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
	Project      string     `gorm:"size:100;uniqueIndex:resource_idx" valid:"length(0|255)"`
	Domain       string     `gorm:"size:100;uniqueIndex:resource_idx" valid:"length(0|255)"`
	Workflow     string     `gorm:"size:100;uniqueIndex:resource_idx" valid:"length(0|255)"`
	LaunchPlan   string     `gorm:"size:100;uniqueIndex:resource_idx" valid:"length(0|255)"`
	ResourceType string     `gorm:"size:100;uniqueIndex:resource_idx" valid:"length(0|255)"`
	Priority     ResourcePriority
	// Serialized flyteidl.admin.MatchingAttributes.
	Attributes []byte
}
