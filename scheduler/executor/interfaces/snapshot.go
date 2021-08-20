package interfaces

import "time"

// Snapshoter used by the scheduler for creating, updating and reading snapshots of the schedules.
type Snapshoter interface {
	// GetLastExecutionTime of the schedule given by the key
	GetLastExecutionTime(key string) time.Time
	// UpdateLastExecutionTime of the schedule given by key to the lastExecTime
	UpdateLastExecutionTime(key string, lastExecTime time.Time)
	// CreateSnapshot creates the snapshot of all the schedules and there execution times.
	CreateSnapshot() ([]byte, error)
	// BootstrapFrom bootstraps the snapshot from a byte array
	BootstrapFrom(snapshot []byte) error
	// GetVersion gets the version number of snapshot written
	GetVersion() int
	// IsEmpty returns true if the snapshot contains no schedules
	IsEmpty() bool
	// AreEqual returns true if the prevSnapshot equals the schedules of the current snapshot
	AreEqual(prevSnapshot Snapshoter) bool
	// Clone returns copy of the current in memory snapshot
	Clone() Snapshoter
}
