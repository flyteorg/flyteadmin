package models

import (
	"time"
)

type NodeExecutionEvent struct {
	BaseModel
	NodeExecutionKey
	RequestID  string `gorm:"size:255"`
	OccurredAt time.Time
	Phase      string `gorm:"size:255;primary_key"`
}
