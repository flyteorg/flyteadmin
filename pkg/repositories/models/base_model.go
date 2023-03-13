package models

import "time"

// This is the base model definition every flyteadmin model embeds.
// This is nearly identical to http://doc.gorm.io/models.html#conventions except that flyteadmin models define their
// own primary keys rather than use the ID as the primary key
type BaseModel struct {
	ID        uint `gorm:"index;autoIncrement"`
	CreatedAt time.Time `gorm:"type:time"`
	UpdatedAt time.Time `gorm:"type:time"`
	DeletedAt *time.Time `gorm:"index"`
}
