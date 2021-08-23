package executor

import (
	"bytes"
	"context"
	"fmt"
	"github.com/flyteorg/flyteadmin/scheduler/executor/interfaces"
	"github.com/flyteorg/flyteadmin/scheduler/repositories"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/wait"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"
)


const snapShotVersion = 1

type checkPointerMetrics struct {
	Scope                        promutils.Scope
	CheckPointPanicCounter       prometheus.Counter
	CheckPointSaveErrCounter     prometheus.Counter
	CheckPointCreationErrCounter prometheus.Counter
}

type ScheduleCheckPointer struct {
	metrics              checkPointerMetrics
	snapshoter           interfaces.Snapshoter
	snapShotReaderWriter interfaces.SnapshotReaderWriter
	db                   repositories.SchedulerRepoInterface
}

func (w *ScheduleCheckPointer) RunCheckPointer(ctx context.Context) {

	checkPointerCtxWithLabel := contextutils.WithGoroutineLabel(ctx, checkPointerRoutineLabel)
	go func(ctx context.Context) {
		pprof.SetGoroutineLabels(ctx)
		defer func() {
			if err := recover(); err != nil {
				w.metrics.CheckPointPanicCounter.Inc()
				logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
			}
		}()
		wait.UntilWithContext(ctx, w.CheckPointState, snapshotWriterSleepTime*time.Second)
	}(checkPointerCtxWithLabel)
}

func (w *ScheduleCheckPointer) CheckPointState(ctx context.Context) {
	var bytesArray []byte
	f := bytes.NewBuffer(bytesArray)
	// Only write if the snapshot has contents and not equal to the previous snapshot
	if !w.snapshoter.IsEmpty() {
		err := w.snapShotReaderWriter.WriteSnapshot(f, w.snapshoter)
		// Just log the error
		if err != nil {
			w.metrics.CheckPointCreationErrCounter.Inc()
			logger.Errorf(ctx, "unable to write the snapshot to buffer due to %v", err)
		}
		err = w.db.ScheduleEntitiesSnapshotRepo().CreateSnapShot(ctx, models.ScheduleEntitiesSnapshot{
			Snapshot: f.Bytes(),
		})
		if err != nil {
			w.metrics.CheckPointSaveErrCounter.Inc()
			logger.Errorf(ctx, "unable to save the snapshot to the database due to %v", err)
		}
	}
}

func (w *ScheduleCheckPointer) ReadCheckPoint(ctx context.Context) interfaces.Snapshoter {
	var snapshot interfaces.Snapshoter
	scheduleEntitiesSnapShot, err := w.db.ScheduleEntitiesSnapshotRepo().GetLatestSnapShot(ctx)

	w.snapshoter = &SnapshotV1{LastTimes: sync.Map{}}
	// Just log the error but dont interrupt the startup of the scheduler
	if err != nil {
		logger.Errorf(ctx, "unable to read the snapshot from the DB due to %v", err)
	} else {
		f := bytes.NewReader(scheduleEntitiesSnapShot.Snapshot)
		snapShotReaderWriter := VersionedSnapshot{version: snapShotVersion}
		snapshot, err = snapShotReaderWriter.ReadSnapshot(f)
		// Similarly just log the error but dont interrupt the startup of the scheduler
		if err != nil {
			logger.Errorf(ctx, "unable to construct the snapshot struct from the file due to %v", err)
		}
		w.snapshoter = snapshot
	}
	return w.snapshoter
}

func NewScheduleCheckPointer(scope promutils.Scope, snapShotReaderWriter interfaces.SnapshotReaderWriter,
	db repositories.SchedulerRepoInterface) interfaces.ScheduleCheckPointer {

	return &ScheduleCheckPointer{
		metrics:         getCheckPointerMetrics(scope),
		snapShotReaderWriter: snapShotReaderWriter,
		db:              db,
	}
}

func getCheckPointerMetrics(scope promutils.Scope) checkPointerMetrics {
	return checkPointerMetrics{
		Scope: scope,
		CheckPointPanicCounter: scope.MustNewCounter("checkpoint_panic_counter",
			"count of crashes for the checkpointer"),
		CheckPointSaveErrCounter: scope.MustNewCounter("checkpoint_save_error_counter",
			"count of unsuccessful attempts to save the created snapshot to the DB"),
		CheckPointCreationErrCounter: scope.MustNewCounter("checkpoint_creation_error_counter",
			"count of unsuccessful attempts to create the snapshot from the inmemory map"),
	}
}