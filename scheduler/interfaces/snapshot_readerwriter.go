package interfaces

import "io"

// SnapshotReaderWriter provides an interface to read and write the snapshot
type SnapshotReaderWriter interface {
	// ReadSnapshot reads the snapshot from the reader
	ReadSnapshot(reader io.Reader) (Snapshoter, error)
	// WriteSnapshot writes the snapshot to the writer
	WriteSnapshot(writer io.Writer, snapshot Snapshoter) error
}
