package scheduler

import (
	"bytes"
	"encoding/gob"
	interfaces2 "github.com/flyteorg/flyteadmin/scheduler/interfaces"
	"reflect"
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
	LastTimes map[string]time.Time
}

func (s *SnapshotV1) GetLastExecutionTime(key string) time.Time {
	return s.LastTimes[key]
}

func (s *SnapshotV1) UpdateLastExecutionTime(key string, lastExecTime time.Time) {
	s.LastTimes[key] = lastExecTime
}

func (s *SnapshotV1) CreateSnapshot() ([]byte, error) {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(s)
	return b.Bytes(), err
}

func (s *SnapshotV1) BootstrapFrom(snapshot []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(snapshot)).Decode(s)
}

func (s *SnapshotV1) IsEmpty() bool {
	return len(s.LastTimes) == 0
}

func (s *SnapshotV1) AreEqual(prevSnapshot interfaces2.Snapshoter) bool {
	if prevSnapshotV1, ok := prevSnapshot.(*SnapshotV1); !ok {
		return false
	} else {
		return reflect.DeepEqual(s.LastTimes, prevSnapshotV1.LastTimes)
	}
}

// Clone returns copy of the current in memory snapshot
func (s *SnapshotV1) Clone() interfaces2.Snapshoter {
	cloned := map[string]time.Time{}
	for k,v := range s.LastTimes {
		cloned[k] = v
	}
	return &SnapshotV1{cloned}
}

func (s *SnapshotV1) GetVersion() int {
	return 1
}
