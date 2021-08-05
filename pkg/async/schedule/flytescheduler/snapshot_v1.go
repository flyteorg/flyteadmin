package flytescheduler

import (
	"bytes"
	"encoding/gob"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/flytescheduler/interfaces"
	"reflect"
	"time"
)

type SnapshotV1 struct {
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

func (s *SnapshotV1) AreEqual(prevSnapshot interfaces.Snapshot) bool {
	if prevSnapshotV1, ok := prevSnapshot.(*SnapshotV1); !ok {
		return false
	} else {
		return reflect.DeepEqual(s.LastTimes, prevSnapshotV1.LastTimes)
	}
}

func (s *SnapshotV1) GetVersion() int {
	return 1
}
