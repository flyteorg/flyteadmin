package interfaces

import "io"

type SnapshotReaderWriter interface {
	ReadSnapshot(reader io.Reader) (Snapshot, error)
	WriteSnapshot(writer io.Writer, snapshot Snapshot) error
}
