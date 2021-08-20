package executor

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/google/uuid"
	"hash/fnv"
	"strconv"
	"time"
)

// Utility functions used by the flyte native scheduler

const (
	scheduleNameInputsFormat = "%s:%s:%s:%s"
	executionIdInputsFormat = scheduleNameInputsFormat +":%s"
)

// GetScheduleName generate the schedule name to be used as unique identification string within the scheduler
func GetScheduleName(s models.SchedulableEntity) string {
	return strconv.FormatUint(hashIdentifier(core.Identifier{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    s.Name,
		Version: s.Version,
	}), 10)
}

// GetExecutionIdentifier returns UUID using the hashed value of the schedule identifier and the scheduledTime
func GetExecutionIdentifier(identifier core.Identifier, scheduledTime time.Time) (uuid.UUID, error) {
	hashValue := hashScheduledTimeStamp(identifier, scheduledTime)
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, hashValue)
	return uuid.FromBytes(b)
}

// hashIdentifier returns the hash of the identifier
func hashIdentifier(identifier core.Identifier) uint64 {
	h := fnv.New64()
	_, err := h.Write([]byte(fmt.Sprintf(scheduleNameInputsFormat,
		identifier.Project, identifier.Domain, identifier.Name, identifier.Version)))
	if err != nil {
		// This shouldn't occur.
		logger.Errorf(context.Background(),
			"failed to hash launch plan identifier: %+v to get schedule name with err: %v", identifier, err)
		return 0
	}
	logger.Debugf(context.Background(), "Returning hash for [%+v]: %d", identifier, h.Sum64())
	return h.Sum64()
}

// hashScheduledTimeStamp return the hash of the identifier and the scheduledTime
func hashScheduledTimeStamp(identifier core.Identifier, scheduledTime time.Time) uint64 {
	h := fnv.New64()
	_, err := h.Write([]byte(fmt.Sprintf(executionIdInputsFormat,
		identifier.Project, identifier.Domain, identifier.Name, identifier.Version, scheduledTime)))
	if err != nil {
		// This shouldn't occur.
		logger.Errorf(context.Background(),
			"failed to hash launch plan identifier: %+v  with scheduled time %v to get execution identifier with err: %v", identifier, scheduledTime, err)
		return 0
	}
	logger.Debugf(context.Background(), "Returning hash for [%+v] %v: %d", identifier, scheduledTime, h.Sum64())
	return h.Sum64()
}
