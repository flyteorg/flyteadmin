package models

import (
	"time"
)

type NodeExecutionEvent struct {
	BaseModel
	NodeExecutionKey
	RequestID  string
	OccurredAt time.Time
	Phase      string `gorm:"size:128;primary_key"`
}
