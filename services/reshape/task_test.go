package reshape_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskToTaskInfo_Timing(t *testing.T) {
	wp := task.NewTestWorkerPool(2)
	wp.Run(task.NewTestContext())

	wp.RegisterJob("reshape-timing-job", func(tsk *task.Task) {
		tsk.Success()
	})

	meta := task.NewTestMetadata("reshape-timing-job")
	tsk, err := wp.DispatchJob(context.Background(), "reshape-timing-job", meta, nil)
	require.NoError(t, err)

	tsk.Wait()

	info := reshape.TaskToTaskInfo(tsk)

	assert.Equal(t, tsk.GetStartTime(), info.StartTime)
	assert.Equal(t, tsk.GetQueueTime(), info.QueueTime)
	assert.Equal(t, tsk.GetFinishTime(), info.FinishTime)
	assert.False(t, info.FinishTime.IsZero(), "finish time should be set for a completed task")
	assert.False(t, info.FinishTime.Before(info.StartTime), "finish time should not precede start time")
	assert.Equal(t, tsk.GetMeta().FormatToResult().ToMap(), info.Metadata, "metadata should be carried onto the task info")
}

func TestTaskToTaskInfo_Error(t *testing.T) {
	wp := task.NewTestWorkerPool(2)
	wp.Run(task.NewTestContext())

	wp.RegisterJob("reshape-error-job", func(tsk *task.Task) {
		tsk.Fail(errors.New("boom: something broke"))
	})

	meta := task.NewTestMetadata("reshape-error-job")
	tsk, err := wp.DispatchJob(context.Background(), "reshape-error-job", meta, nil)
	require.NoError(t, err)

	tsk.Wait()

	info := reshape.TaskToTaskInfo(tsk)

	assert.Equal(t, "error", info.Status)
	assert.Contains(t, info.Error, "boom: something broke", "the failing task's error message should be on the task info")
}
