package core

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/stretchr/testify/assert"
)

func TestGetCronScheduledTime(t *testing.T) {
	fromTime := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
	nextTime, err := getCronScheduledTime("0 19 * * *", fromTime)
	assert.Nil(t, err)
	expectedNextTime := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedNextTime, nextTime)
}

func TestGetCatchUpTimes(t *testing.T) {
	t.Run("to time before scheduled time", func(t *testing.T) {
		s := models.SchedulableEntity{
			CronExpression: "0 19 * * *",
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 00, 51, 6, 0, time.UTC)
		catchupTimes, err := GetCatchUpTimes(s, from, to)
		assert.Nil(t, err)
		assert.Nil(t, catchupTimes)
	})
	t.Run("to time equal scheduled time", func(t *testing.T) {
		s := models.SchedulableEntity{
			CronExpression: "0 19 * * *",
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
		catchupTimes, err := GetCatchUpTimes(s, from, to)
		assert.Nil(t, err)
		assert.NotNil(t, catchupTimes)
		assert.Equal(t, 1, len(catchupTimes))
		expectedNextTime := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedNextTime, catchupTimes[0])
	})
	t.Run("to time after scheduled time", func(t *testing.T) {
		s := models.SchedulableEntity{
			CronExpression: "0 19 * * *",
		}
		from := time.Date(2022, time.January, 27, 19, 0, 0, 0, time.UTC)
		to := time.Date(2022, time.January, 28, 19, 50, 0, 0, time.UTC)
		catchupTimes, err := GetCatchUpTimes(s, from, to)
		assert.Nil(t, err)
		assert.NotNil(t, catchupTimes)
		assert.Equal(t, 1, len(catchupTimes))
		expectedNextTime := time.Date(2022, time.January, 28, 19, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedNextTime, catchupTimes[0])
	})
}
