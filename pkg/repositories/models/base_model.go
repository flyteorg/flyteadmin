package models

import "time"

// This is the base model definition every flyteadmin model embeds.
// This is nearly identical to http://doc.gorm.io/models.html#conventions except that flyteadmin models define their
// own primary keys rather than use the ID as the primary key
type BaseModel struct {
	ID        *int32 `gorm:"index;AUTOINCREMENT"`
	//ID        uint `gorm:"type:int;index;AUTOINCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
