package task

import (
	"math/rand"
	"testing"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var jobName = "Test Job"
var subpoolJobName = "Test Subpool Job"

func TestWorkerPoolBasic(t *testing.T) {
	t.Parallel()

	wpLogger := zerolog.Nop()
	wp := NewWorkerPool(2, &wpLogger)
	wp.RegisterJob(jobName, testJob)

	wp.Run()
	defer wp.Stop()

	tsk, err := wp.DispatchJob(jobName, fakeJobMeta{}, nil)
	require.NoError(t, err)

	tsk.Wait()
	tskResult := tsk.GetResult("test")
	assert.NotNil(t, tskResult)
	assert.Equal(t, "passed", tskResult.(string))
}

func TestSubPool(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	t.Parallel()

	wpLogger := zerolog.Nop()
	wp := NewWorkerPool(2, &wpLogger)
	wp.RegisterJob(jobName, testJob)
	wp.RegisterJob(subpoolJobName, testSubpoolJob)
	wp.Run()

	tsk, err := wp.DispatchJob(
		subpoolJobName, fakeJobMeta{
			subTaskCount: rand.Int() % 1024,
		}, nil,
	)
	require.NoError(t, err)

	tsk.Wait()
	tskResult := tsk.GetResult("test")
	assert.NotNil(t, tskResult)
	assert.Equal(t, "passed", tskResult.(string))

	childPool := tsk.GetChildTaskPool()
	assert.NotNil(t, childPool)
	assert.Equal(t, tsk, childPool.CreatedInTask())

	wp.Stop()

	_, _, _, workers, _ := wp.Status()
	assert.Equal(t, 0, workers)
}

func TestFailedJob(t *testing.T) {
	t.Parallel()

	wpLogger := zerolog.Nop()
	wp := NewWorkerPool(2, &wpLogger)
	wp.RegisterJob(jobName, testJob)
	wp.RegisterJob(subpoolJobName, testSubpoolJob)
	wp.Run()
	defer wp.Stop()

	tsk, err := wp.DispatchJob(jobName, fakeJobMeta{shouldFail: true}, nil)
	require.NoError(t, err)
	tsk.Wait()

	_, exitStatus := tsk.Status()
	assert.Equal(t, TaskError, exitStatus)
}

type fakeJobMeta struct {
	taskNum      int
	subTaskCount int
	shouldFail   bool
}

func (f fakeJobMeta) JobName() string {
	if f.subTaskCount == 0 {
		return jobName
	} else {
		return subpoolJobName
	}
}

func (f fakeJobMeta) MetaString() string {
	return jobName
}

func (f fakeJobMeta) FormatToResult() TaskResult {
	return TaskResult{}
}

func (f fakeJobMeta) Verify() error {
	return nil
}

func testJob(t *Task) {
	meta := t.GetMeta().(fakeJobMeta)
	if meta.shouldFail {
		t.ReqNoErr(werror.Errorf("test error"))
	}

	t.SetResult(TaskResult{"test": "passed", "taskNum": meta.taskNum})
	t.Success()
}

func testSubpoolJob(t *Task) {
	meta := t.GetMeta().(fakeJobMeta)

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(false, t)
	t.SetChildTaskPool(pool)

	for i := range meta.subTaskCount {
		newMeta := fakeJobMeta{taskNum: i}
		_, err := pool.GetWorkerPool().DispatchJob(jobName, newMeta, pool)
		if err != nil {
			t.ReqNoErr(err)
		}
	}

	pool.SignalAllQueued()
	pool.Wait(false)

	t.SetResult(TaskResult{"test": "passed"})
	t.Success()
}
