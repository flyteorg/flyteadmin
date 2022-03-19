package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NodeExecutionEvent struct {
	BaseModel
	NodeExecutionKey
	RequestID  string
	OccurredAt time.Time
	Phase      string `gorm:"primary_key"`
}

func (n *NodeExecutionEvent) BeforeCreate(tx *gorm.DB) error {
	n.ID = uint(uuid.New().ID())
	return nil
}
