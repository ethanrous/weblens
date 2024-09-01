package task_test

import (
	"math/rand"
	"testing"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	. "github.com/ethrousseau/weblens/task"
	"github.com/stretchr/testify/assert"
)

var jobName = "Test Job"
var subpoolJobName = "Test Subpool Job"

func init() {
	internal.IsDevMode()
}

func TestWorkerPoolBasic(t *testing.T) {
	t.Parallel()

	wp := NewWorkerPool(4, -1)
	wp.RegisterJob(jobName, testJob)
	wp.Run()

	tsk, err := wp.DispatchJob(jobName, fakeJobMeta{}, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	tsk.Wait()
	tskResult := tsk.GetResult("test")
	assert.NotNil(t, tskResult)
	assert.Equal(t, "passed", tskResult.(string))

	wp.Stop()
}

func TestSubPool(t *testing.T) {
	t.Parallel()

	wp := NewWorkerPool(4, -1)
	wp.RegisterJob(jobName, testJob)
	wp.RegisterJob(subpoolJobName, testSubpoolJob)
	wp.Run()

	tsk, err := wp.DispatchJob(
		subpoolJobName, fakeJobMeta{
			subTaskCount: rand.Int() % 64,
		}, nil,
	)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

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

	wp := NewWorkerPool(4, -1)
	wp.RegisterJob(jobName, testJob)
	wp.RegisterJob(subpoolJobName, testSubpoolJob)
	wp.Run()

	tsk, err := wp.DispatchJob(jobName, fakeJobMeta{shouldFail: true}, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
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
		t.ErrorAndExit(werror.Errorf("oh no"))
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
			t.ErrorAndExit(err)
		}
	}

	pool.SignalAllQueued()
	pool.Wait(false)

	t.SetResult(TaskResult{"test": "passed"})
	t.Success()
}
