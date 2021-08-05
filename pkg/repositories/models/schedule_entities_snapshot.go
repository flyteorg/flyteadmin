package models

// Database model to save the snapshot for the schedulable entities in the db
type ScheduleEntitiesSnapshot struct {
	BaseModel
	Snapshot []byte `gorm:"column:snapshot" schema:"-"`
}

type ScheduleEntitiesSnapshotCollectionOutput struct {
	Snapshots []ScheduleEntitiesSnapshot
}
