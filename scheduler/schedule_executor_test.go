// +build !race

package scheduler

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	adminModels "github.com/flyteorg/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	schedMocks "github.com/flyteorg/flyteadmin/scheduler/repositories/mocks"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteadmin/scheduler/snapshoter"
	adminMocks "github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var schedules []models.SchedulableEntity

func setupScheduleExecutor(t *testing.T) ScheduledExecutor {
	db := mocks.NewMockRepository()
	var scope = promutils.NewScope("test_scheduler")
	scheduleExecutorConfig := runtimeInterfaces.WorkflowExecutorConfig{
		FlyteWorkflowExecutorConfig: &runtimeInterfaces.FlyteWorkflowExecutorConfig{
			AdminRateLimit: &runtimeInterfaces.AdminRateLimit{
				Tps: 100,
				Burst: 10,
			},
		},
	}
	var bytesArray []byte
	f := bytes.NewBuffer(bytesArray)
	writer := snapshoter.VersionedSnapshot{}
	snapshot := &snapshoter.SnapshotV1{
		LastTimes: map[string]*time.Time{},
	}
	err := writer.WriteSnapshot(f, snapshot)
	assert.Nil(t, err)
	mockAdminClient := new(adminMocks.AdminServiceClient)
	snapshotRepo := db.ScheduleEntitiesSnapshotRepo().(*schedMocks.ScheduleEntitiesSnapShotRepoInterface)

	scheduleEntitiesRepo := db.SchedulableEntityRepo().(*schedMocks.SchedulableEntityRepoInterface)

	snapshotModel := models.ScheduleEntitiesSnapshot{
		BaseModel: adminModels.BaseModel{
			ID:        17,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Snapshot: f.Bytes(),
	}

	activeV2 := true
	schedules = append(schedules, models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron_schedule",
			Version: "v2",
		},
		CronExpression:      "@every 1s",
		KickoffTimeInputArg: "kickoff_time",
		Active:              &activeV2,
	})

	activeV1 := false
	schedules = append(schedules, models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron_schedule",
			Version: "v1",
		},
		CronExpression:      "*/1 * * * *",
		KickoffTimeInputArg: "kickoff_time",
		Active:              &activeV1,
	})
	snapshotRepo.OnRead(context.Background()).Return(snapshotModel, nil)
	scheduleEntitiesRepo.OnGetAllMatch(mock.Anything).Return(schedules, nil)
	mockAdminClient.OnCreateExecutionMatch(context.Background(), mock.Anything).
		Return(&admin.ExecutionCreateResponse{}, nil)

	return NewScheduledExecutor(db, scheduleExecutorConfig,
		scope, mockAdminClient)
}

func TestSuccessfulSchedulerExec(t *testing.T) {
	t.Run("add cron schedule", func(t *testing.T) {
		scheduleExecutor := setupScheduleExecutor(t)

		go func() {
			err := scheduleExecutor.Run(context.Background())
			assert.Nil(t, err)
		}()
		time.Sleep(5 * time.Second)
	})
}
