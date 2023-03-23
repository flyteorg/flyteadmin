package models

// Signal primary key
type SignalKey struct {
	ExecutionKey
	SignalID string `gorm:"size:255;primary_key;index"`
}

// Database model to encapsulate a signal.
type Signal struct {
	BaseModel
	SignalKey
	Type  []byte `gorm:"not null"`
	Value []byte
}
