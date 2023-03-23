package models

import (
	"time"
)

type ExecutionEvent struct {
	BaseModel
	ExecutionKey
	RequestID  string `gorm:"size:255"`
	OccurredAt time.Time
	Phase      string `gorm:"size:255;primary_key"`
}
