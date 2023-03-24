package models

import (
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Database model to encapsulate metadata associated with a SchedulableEntity
type SchedulableEntity struct {
	models.BaseModel
	SchedulableEntityKey
	// FIXME: figure out if this is just the schedule definition.
	CronExpression      string `gorm:"size:100"`
	FixedRateValue      uint32
	Unit                admin.FixedRateUnit
	// FIXME: figure out how big this should be.
	KickoffTimeInputArg string `gorm:"size:100"`
	Active              *bool
}

// Schedulable entity primary key
type SchedulableEntityKey struct {
	Project string `gorm:"size:64;primary_key"`
	Domain  string `gorm:"size:255;primary_key"`
	Name    string `gorm:"size:255;primary_key"`
	Version string `gorm:"size:255;primary_key"`
}
