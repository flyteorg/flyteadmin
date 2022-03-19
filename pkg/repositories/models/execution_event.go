package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExecutionEvent struct {
	BaseModel
	ExecutionKey
	RequestID  string `valid:"length(0|255)"`
	OccurredAt time.Time
	Phase      string `gorm:"primary_key"`
}

func (e *ExecutionEvent) BeforeCreate(tx *gorm.DB) error {
	e.ID = uint(uuid.New().ID())
	return nil
}
