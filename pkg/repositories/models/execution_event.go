package models

import (
	"time"
)

type ExecutionEvent struct {
	BaseModel
	ExecutionKey
	RequestID  string `valid:"length(1|200)"`
	OccurredAt time.Time
	Phase      string `gorm:"primary_key" valid:"length(1|200)"`
}
