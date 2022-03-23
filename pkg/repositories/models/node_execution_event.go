package models

import (
	"time"
)

type NodeExecutionEvent struct {
	BaseModel
	NodeExecutionKey
	RequestID  string
	OccurredAt time.Time
	Phase      string `gorm:"uniqueIndex:primary_nee_index"`
}
