package snapshoter

import (
	"bytes"
	"encoding/gob"
	"time"
)

// snapshotV1 stores in the inmemory states of the schedules and there last execution timestamps.
// This map is created periodically from the jobstore of the gocron_wrapper and written to the DB.
// During bootup the serialized version of it is read from the DB and the schedules are initialized from it.
// V1 version so that in future if we add more fields for extending the functionality then this provides
// a backward compatible way to read old snapshots.
type snapshotV1 struct {
	// LastTimes map of the schedule name to last execution timestamp
	LastTimes map[string]*time.Time
}

func (s *snapshotV1) GetLastExecutionTime(key string) *time.Time {
	return s.LastTimes[key]
}

func (s *snapshotV1) UpdateLastExecutionTime(key string, lastExecTime *time.Time) {
	// Load the last exec time for the schedule key and compare if its less than new LastExecTime
	// and only if it is then update the map
	s.LastTimes[key] = lastExecTime
}

func (s *snapshotV1) Serialize() ([]byte, error) {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(s)
	return b.Bytes(), err
}

func (s *snapshotV1) Deserialize(snapshot []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(snapshot)).Decode(s)
}

func (s *snapshotV1) IsEmpty() bool {
	return len(s.LastTimes) == 0
}

func (s *snapshotV1) GetVersion() int {
	return 1
}

func (s *snapshotV1) Create() Snapshot {
	return &snapshotV1{
		LastTimes: map[string]*time.Time{},
	}
}
