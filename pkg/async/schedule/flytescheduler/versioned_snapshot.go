package flytescheduler

import (
	"encoding/gob"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/flytescheduler/interfaces"
	"io"
	"time"
)

type VersionedSnapshot struct {
	version int
	Ser     []byte
}

func (s *VersionedSnapshot) WriteSnapshot(w io.Writer, snapshot interfaces.Snapshoter) error {
	byteContents, err := snapshot.CreateSnapshot()
	if err != nil {
		return err
	}
	s.version = snapshot.GetVersion()
	s.Ser = byteContents
	enc := gob.NewEncoder(w)
	return enc.Encode(s)
}

func (s *VersionedSnapshot) ReadSnapshot(r io.Reader) (interfaces.Snapshoter, error) {
	err := gob.NewDecoder(r).Decode(s)
	if err != nil {
		return nil, err
	}
	if s.version == 1 {
		snapShotV1 := SnapshotV1{LastTimes: map[string]time.Time{}}
		err = snapShotV1.BootstrapFrom(s.Ser)
		if err != nil {
			return nil, err
		}
		return &snapShotV1, nil
	}
	return nil, fmt.Errorf("unsupported version %v", s.version)
}
