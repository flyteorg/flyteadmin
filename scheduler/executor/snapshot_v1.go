package executor

import (
	"bytes"
	"encoding/gob"
	"sync"
	"time"
)

// SnapshotV1 stores in the inmemory states of the schedules and there last execution timestamps.
// This map is continuously updated after each execution is successfully sent for execution to admin
// Inorder for this map to survive crashes the SnapshotReaderWriter interface provides a way to persist and bootstrap
// from the persisted instance of the snapshot
// V1 version so that in future if we add more fields for extending the functionality then this provides
// a backward compatible way to read old snapshots.
type SnapshotV1 struct {
	// LastTimes map of the schedule name to last execution timestamp
	LastTimes sync.Map
}

// lastExecTimeEntry This is only used for gob serialization and deserialization since sync.Map cannot do the same.
type lastExecTimeEntry struct {
	LastExecTime       time.Time
	ScheduleIdentifier string
}

func (s *SnapshotV1) GetLastExecutionTime(key string) time.Time {
	val, ok := s.LastTimes.Load(key)
	if !ok {
		return time.Time{}
	}
	return val.(time.Time)
}

func (s *SnapshotV1) UpdateLastExecutionTime(key string, lastExecTime time.Time) {
	// Load the last exec time for the schedule key and compare if its less than new LastExecTime
	// and only if it is then update the map
	prevLastExecTime, ok := s.LastTimes.Load(key)
	if !ok || prevLastExecTime.(time.Time).Before(lastExecTime){
		s.LastTimes.Store(key, lastExecTime)
	}
}

func (s *SnapshotV1) CreateSnapshot() ([]byte, error) {
	var b bytes.Buffer
	lastExecTimeEntries := s.ReadEntries()
	err := gob.NewEncoder(&b).Encode(lastExecTimeEntries)
	return b.Bytes(), err
}

func (s *SnapshotV1) ReadEntries() []lastExecTimeEntry {
	var lastExecTimeEntries []lastExecTimeEntry
	s.LastTimes.Range(func(key, value interface{}) bool {
		lastExecTime := value.(time.Time)
		scheduleIdentifier := key.(string)
		lastExecTimeEntries = append(lastExecTimeEntries, lastExecTimeEntry{
			LastExecTime: lastExecTime, ScheduleIdentifier: scheduleIdentifier,
		})
		return true
	})
	return lastExecTimeEntries
}

func (s *SnapshotV1) BootstrapFrom(snapshot []byte) error {
	var lastExecTimeEntries [] lastExecTimeEntry
	err := gob.NewDecoder(bytes.NewBuffer(snapshot)).Decode(&lastExecTimeEntries)
	if err != nil {
		return err
	}
	for _, entry := range lastExecTimeEntries {
		s.LastTimes.Store(entry.ScheduleIdentifier, entry.LastExecTime)
	}
	return nil
}

func (s *SnapshotV1) IsEmpty() bool {
	lastExecTimeEntries := s.ReadEntries()
	return len(lastExecTimeEntries) == 0
}

func (s *SnapshotV1) GetVersion() int {
	return 1
}
