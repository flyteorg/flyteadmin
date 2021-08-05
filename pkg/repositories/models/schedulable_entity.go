package models

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// Database model to encapsulate metadata associated with a SchedulableEntity
type SchedulableEntity struct {
	BaseModel
	SchedulableEntityKey
	CronExpression      string
	FixedRateValue      uint32
	Unit                admin.FixedRateUnit
	KickoffTimeInputArg string
	Active              *bool
}

type SchedulableEntityCollectionOutput struct {
	Entities []SchedulableEntity
}

// Schedulable entity primary key
type SchedulableEntityKey struct {
	Project string `gorm:"primary_key"`
	Domain  string `gorm:"primary_key"`
	Name    string `gorm:"primary_key"`
	Version string `gorm:"primary_key"`
}
