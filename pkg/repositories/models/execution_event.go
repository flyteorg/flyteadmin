package models

import (
	"time"
)

type ExecutionEvent struct {
	BaseModel
	ExecutionKey
	RequestID  string `valid:"length(1|50)"`
	OccurredAt time.Time
	Phase      string `gorm:"primary_key" valid:"length(1|50)"`
}
