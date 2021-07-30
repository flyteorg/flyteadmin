package models

import (
	"time"
)

// Database model to encapsulate metadata associated with a SchedulableEntity
type ScheduleCheckPoint struct {
	CheckPointTime   *time.Time
}

