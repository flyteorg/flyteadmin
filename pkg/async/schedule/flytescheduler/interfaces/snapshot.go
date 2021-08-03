package interfaces

import "time"

// All threadsafe
type Snapshot interface {
	GetLastExecutionTime(key string) time.Time
	UpdateLastExecutionTime(key string, lastExecTime time.Time)
	CreateSnapshot() ([]byte, error)
	BootstrapFrom(snapshot []byte) error
	GetVersion() int
}
