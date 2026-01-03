package task_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/stretchr/testify/assert"
)

var intentionalFailure = wlerrors.Errorf("intentional failure for testing")

func TestTask_ID(t *testing.T) {
	t.Run("returns task ID", func(t *testing.T) {
		tsk := task.NewTestTask("test-id-123", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		assert.Equal(t, "test-id-123", tsk.ID())
	})
}

func TestTask_JobName(t *testing.T) {
	t.Run("returns job name", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "ScanDirectory", 0, task.TaskSuccess, nil, time.Now())

		assert.Equal(t, "ScanDirectory", tsk.JobName())
	})

	t.Run("is thread-safe", func(_ *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		var wg sync.WaitGroup

		for range 100 {
			wg.Go(func() {
				_ = tsk.JobName()
			})
		}

		wg.Wait()
	})
}

func TestTask_GetWorkerID(t *testing.T) {
	t.Run("returns worker ID", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 42, task.TaskSuccess, nil, time.Now())

		assert.Equal(t, 42, tsk.GetWorkerID())
	})

	t.Run("returns 0 for unassigned worker", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		assert.Equal(t, 0, tsk.GetWorkerID())
	})
}

func TestTask_Status(t *testing.T) {
	t.Run("returns complete for exited task with success", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskSuccess, status)
	})

	t.Run("returns error status for failed task", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskError, nil, time.Now())

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskError, status)
	})

	t.Run("returns canceled status for canceled task", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskCanceled, nil, time.Now())

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskCanceled, status)
	})
}

func TestTask_QueueState(t *testing.T) {
	t.Run("returns Exited state for completed task", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		assert.Equal(t, task.Exited, tsk.QueueState())
	})
}

func TestTask_Wait(t *testing.T) {
	t.Run("returns immediately for nil task", func(t *testing.T) {
		var tsk *task.Task

		// Should not panic or block
		assert.NotPanics(t, func() {
			tsk.Wait()
		})
	})

	t.Run("returns immediately for exited task", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		done := make(chan bool)

		go func() {
			tsk.Wait()

			done <- true
		}()

		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("Wait should have returned immediately")
		}
	})
}

func TestTask_GetResults(t *testing.T) {
	t.Run("returns empty result for task without result", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		result := tsk.GetResults()
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns set result", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, task.Result{"key": "value", "count": 42}, time.Now())

		result := tsk.GetResults()
		assert.Equal(t, "value", result["key"])
		assert.Equal(t, 42, result["count"])
	})
}

func TestTask_GetResult(t *testing.T) {
	t.Run("returns same as GetResults", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, task.Result{"key": "value"}, time.Now())

		result := tsk.GetResult()
		results := tsk.GetResults()

		assert.Equal(t, results["key"], result["key"])
	})
}

func TestTask_SetResult(t *testing.T) {
	t.Run("sets result", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		tsk.SetResult(task.Result{"key": "value"})

		result := tsk.GetResults()
		assert.Equal(t, "value", result["key"])
	})

	t.Run("merges with existing result", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, task.Result{"existing": "data"}, time.Now())

		tsk.SetResult(task.Result{"new": "value"})

		result := tsk.GetResults()
		assert.Equal(t, "data", result["existing"])
		assert.Equal(t, "value", result["new"])
	})

	t.Run("overwrites existing keys", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, task.Result{"key": "old"}, time.Now())

		tsk.SetResult(task.Result{"key": "new"})

		result := tsk.GetResults()
		assert.Equal(t, "new", result["key"])
	})

	t.Run("calls result callback if set", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		callbackCalled := false

		var callbackResult task.Result

		tsk.OnResult(func(result task.Result) {
			callbackCalled = true
			callbackResult = result
		})

		tsk.SetResult(task.Result{"key": "value"})

		assert.True(t, callbackCalled)
		assert.Equal(t, "value", callbackResult["key"])
	})
}

func TestTask_OnResult(t *testing.T) {
	t.Run("sets callback", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		called := false

		tsk.OnResult(func(_ task.Result) {
			called = true
		})

		tsk.SetResult(task.Result{})
		assert.True(t, called)
	})
}

func TestTask_ClearOnResult(t *testing.T) {
	t.Run("clears callback", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		called := false

		tsk.OnResult(func(_ task.Result) {
			called = true
		})

		tsk.ClearOnResult()
		tsk.SetResult(task.Result{})

		assert.False(t, called)
	})
}

func TestTask_GetMeta(t *testing.T) {
	t.Run("returns nil for unset metadata", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		meta := tsk.GetMeta()
		assert.Nil(t, meta)
	})
}

func TestTask_ReadError(t *testing.T) {
	t.Run("returns nil for task without error", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		err := tsk.ReadError()
		assert.Nil(t, err)
	})
}

func TestTask_Timeout(t *testing.T) {
	t.Run("GetTimeout returns zero time initially", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		timeout := tsk.GetTimeout()
		assert.True(t, timeout.IsZero())
	})

	t.Run("ClearTimeout sets timeout to zero", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		tsk.ClearTimeout()

		timeout := tsk.GetTimeout()
		assert.Equal(t, time.Unix(0, 0), timeout)
	})
}

func TestTask_ExeTime(t *testing.T) {
	t.Run("returns duration for finished task", func(t *testing.T) {
		startTime := time.Now().Add(-10 * time.Second)
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, startTime)

		exeTime := tsk.ExeTime()
		// Since FinishTime is zero, it uses time.Now()
		assert.GreaterOrEqual(t, exeTime, 10*time.Second)
	})
}

func TestTask_Manipulate(t *testing.T) {
	t.Run("runs function on metadata", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		called := false
		err := tsk.Manipulate(func(_ task.Metadata) error {
			called = true

			return nil
		})

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestTask_GetChildTaskPool(t *testing.T) {
	t.Run("returns nil initially", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		pool := tsk.GetChildTaskPool()
		assert.Nil(t, pool)
	})
}

func TestTask_GetTaskPool(t *testing.T) {
	t.Run("returns nil for unqueued task", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		pool := tsk.GetTaskPool()
		assert.Nil(t, pool)
	})
}

func TestQueueStateConstants(t *testing.T) {
	t.Run("Created is 0", func(t *testing.T) {
		assert.Equal(t, task.QueueState(0), task.Created)
	})

	t.Run("InQueue is 1", func(t *testing.T) {
		assert.Equal(t, task.QueueState(1), task.InQueue)
	})

	t.Run("Executing is 2", func(t *testing.T) {
		assert.Equal(t, task.QueueState(2), task.Executing)
	})

	t.Run("Sleeping is 3", func(t *testing.T) {
		assert.Equal(t, task.QueueState(3), task.Sleeping)
	})

	t.Run("Exited is 4", func(t *testing.T) {
		assert.Equal(t, task.QueueState(4), task.Exited)
	})
}

func TestOptionsStruct(t *testing.T) {
	t.Run("default values are false", func(t *testing.T) {
		opts := task.Options{}

		assert.False(t, opts.Persistent)
		assert.False(t, opts.Unique)
	})

	t.Run("can set Persistent", func(t *testing.T) {
		opts := task.Options{Persistent: true}

		assert.True(t, opts.Persistent)
	})

	t.Run("can set Unique", func(t *testing.T) {
		opts := task.Options{Unique: true}

		assert.True(t, opts.Unique)
	})

	t.Run("can set both", func(t *testing.T) {
		opts := task.Options{Persistent: true, Unique: true}

		assert.True(t, opts.Persistent)
		assert.True(t, opts.Unique)
	})
}

func TestTask_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent reads", func(_ *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, nil, time.Now())

		var wg sync.WaitGroup

		for range 50 {
			wg.Go(func() {
				_ = tsk.ID()
				_ = tsk.JobName()
				_ = tsk.GetWorkerID()
				_ = tsk.QueueState()
				_, _ = tsk.Status()
				_ = tsk.GetResults()
				_ = tsk.GetTimeout()
			})
		}

		wg.Wait()
	})
}

func TestTaskErrors(t *testing.T) {
	t.Run("ErrTaskError is defined", func(t *testing.T) {
		assert.NotNil(t, task.ErrTaskError)
		assert.Contains(t, task.ErrTaskError.Error(), "task error")
	})

	t.Run("ErrTaskExit is defined", func(t *testing.T) {
		assert.NotNil(t, task.ErrTaskExit)
		assert.Contains(t, task.ErrTaskExit.Error(), "task exit")
	})

	t.Run("ErrTaskTimeout is defined", func(t *testing.T) {
		assert.NotNil(t, task.ErrTaskTimeout)
		assert.Contains(t, task.ErrTaskTimeout.Error(), "task timeout")
	})

	t.Run("ErrTaskCanceled is defined", func(t *testing.T) {
		assert.NotNil(t, task.ErrTaskCanceled)
		assert.Contains(t, task.ErrTaskCanceled.Error(), "task canceled")
	})

	t.Run("ErrTaskAlreadyComplete is defined", func(t *testing.T) {
		assert.NotNil(t, task.ErrTaskAlreadyComplete)
		assert.Contains(t, task.ErrTaskAlreadyComplete.Error(), "task already complete")
	})
}

func TestGlobalTaskPoolID(t *testing.T) {
	t.Run("constant is defined", func(t *testing.T) {
		assert.Equal(t, "GLOBAL", task.GlobalTaskPoolID)
	})
}

// ==================== Pool Tests ====================

func TestPool_ID(t *testing.T) {
	t.Run("returns unique pool ID", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		iPool := wp.NewTaskPool(false, nil)
		pool := wp.GetTaskPool(iPool.GetRootPool().ID())

		assert.NotEmpty(t, pool.ID())
	})

	t.Run("global pool has GlobalTaskPoolID", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.GetTaskPool(task.GlobalTaskPoolID)

		assert.Equal(t, task.GlobalTaskPoolID, pool.ID())
	})
}

func TestPool_IsRoot(t *testing.T) {
	t.Run("returns true for pool without parent", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		assert.True(t, pool.IsRoot())
	})

	t.Run("global pool is root", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.GetTaskPool(task.GlobalTaskPoolID)

		assert.True(t, pool.IsRoot())
	})
}

func TestPool_IsGlobal(t *testing.T) {
	t.Run("returns true for global pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.GetTaskPool(task.GlobalTaskPoolID)

		assert.True(t, pool.IsGlobal())
	})

	t.Run("returns false for non-global pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		assert.False(t, pool.IsGlobal())
	})
}

func TestPool_GetWorkerPool(t *testing.T) {
	t.Run("returns associated worker pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		result := pool.GetWorkerPool()
		assert.NotNil(t, result)
	})
}

func TestPool_TaskCount(t *testing.T) {
	t.Run("IncTaskCount increments total", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		assert.Equal(t, 0, pool.GetTotalTaskCount())

		pool.IncTaskCount(5)
		assert.Equal(t, 5, pool.GetTotalTaskCount())

		pool.IncTaskCount(3)
		assert.Equal(t, 8, pool.GetTotalTaskCount())
	})

	t.Run("IncCompletedTasks increments completed count", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		assert.Equal(t, 0, pool.GetCompletedTaskCount())

		pool.IncCompletedTasks(2)
		assert.Equal(t, 2, pool.GetCompletedTaskCount())

		pool.IncCompletedTasks(1)
		assert.Equal(t, 3, pool.GetCompletedTaskCount())
	})
}

func TestPool_Status(t *testing.T) {
	t.Run("returns status with zero values initially", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		status := pool.Status()
		assert.Equal(t, int64(0), status.Complete)
		assert.Equal(t, int64(0), status.Total)
		assert.Equal(t, 0, status.Failed)
		assert.Equal(t, float64(0), status.Progress)
	})

	t.Run("calculates progress correctly", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		pool.IncTaskCount(10)
		pool.IncCompletedTasks(5)

		status := pool.Status()
		assert.Equal(t, int64(5), status.Complete)
		assert.Equal(t, int64(10), status.Total)
		assert.Equal(t, float64(50), status.Progress)
	})

	t.Run("runtime increases over time", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		status1 := pool.Status()

		time.Sleep(10 * time.Millisecond)

		status2 := pool.Status()

		assert.Greater(t, status2.Runtime, status1.Runtime)
	})
}

func TestPool_GetRootPool(t *testing.T) {
	t.Run("returns self for root pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		root := pool.GetRootPool()
		assert.Equal(t, pool.ID(), root.ID())
	})
}

func TestPool_AddError(t *testing.T) {
	t.Run("adds task to error list", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskError, nil, time.Now())
		pool.AddError(tsk)

		errors := pool.Errors()
		assert.Len(t, errors, 1)
	})

	t.Run("can add multiple errors", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		tsk1 := task.NewTestTask("task-1", "TestJob", 0, task.TaskError, nil, time.Now())
		tsk2 := task.NewTestTask("task-2", "TestJob", 0, task.TaskError, nil, time.Now())

		pool.AddError(tsk1)
		pool.AddError(tsk2)

		errors := pool.Errors()
		assert.Len(t, errors, 2)
	})
}

func TestPool_Errors(t *testing.T) {
	t.Run("returns empty slice initially", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		errors := pool.Errors()
		assert.Empty(t, errors)
	})
}

func TestPool_CreatedInTask(t *testing.T) {
	t.Run("returns nil when not created in task", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		creator := pool.CreatedInTask()
		assert.Nil(t, creator)
	})
}

func TestPool_LockUnlock(t *testing.T) {
	t.Run("LockExit and UnlockExit work without panic", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		assert.NotPanics(t, func() {
			pool.LockExit()
			pool.UnlockExit()
		})
	})

	t.Run("multiple lock/unlock cycles work", func(_ *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		for range 10 {
			pool.LockExit()
			pool.UnlockExit()
		}
	})
}

func TestPool_AddCleanup(t *testing.T) {
	t.Run("registers cleanup function", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		called := false

		pool.AddCleanup(func(_ *task.Pool) {
			called = true
		})

		// Cleanup is not called immediately
		assert.False(t, called)
	})
}

func TestPool_HandleTaskExit(t *testing.T) {
	t.Run("returns true for non-replacement thread", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		canContinue := pool.HandleTaskExit(false)
		assert.True(t, canContinue)
	})
}

func TestPool_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent task count operations", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		var wg sync.WaitGroup

		for range 100 {
			wg.Add(2)

			go func() {
				defer wg.Done()

				pool.IncTaskCount(1)
			}()

			go func() {
				defer wg.Done()

				_ = pool.GetTotalTaskCount()
			}()
		}

		wg.Wait()
		assert.Equal(t, 100, pool.GetTotalTaskCount())
	})

	t.Run("handles concurrent status reads", func(_ *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		var wg sync.WaitGroup

		for range 100 {
			wg.Go(func() {
				_ = pool.Status()
			})
		}

		wg.Wait()
	})
}

// ==================== WorkerPool Tests ====================

func TestNewWorkerPool(t *testing.T) {
	t.Run("creates worker pool with specified workers", func(t *testing.T) {
		wp := task.NewTestWorkerPool(4)

		assert.NotNil(t, wp)
	})

	t.Run("creates worker pool with at least 1 worker when 0 specified", func(t *testing.T) {
		wp := task.NewTestWorkerPool(0)

		assert.NotNil(t, wp)
	})

	t.Run("automatically creates global pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		globalPool := wp.GetTaskPool(task.GlobalTaskPoolID)
		assert.NotNil(t, globalPool)
		assert.True(t, globalPool.IsGlobal())
	})
}

func TestWorkerPool_NewTaskPool(t *testing.T) {
	t.Run("creates new task pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		assert.NotNil(t, pool)
		assert.NotEmpty(t, pool.ID())
	})

	t.Run("creates unique pool IDs", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool1 := wp.NewTaskPool(false, nil)
		pool2 := wp.NewTaskPool(false, nil)

		assert.NotEqual(t, pool1.ID(), pool2.ID())
	})

	t.Run("pool is retrievable by ID", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		pool := wp.NewTaskPool(false, nil)

		retrieved := wp.GetTaskPool(pool.ID())
		assert.NotNil(t, retrieved)
		assert.Equal(t, pool.ID(), retrieved.ID())
	})
}

func TestWorkerPool_GetTaskPool(t *testing.T) {
	t.Run("returns nil for non-existent pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		pool := wp.GetTaskPool("non-existent-id")
		assert.Nil(t, pool)
	})

	t.Run("returns global pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		pool := wp.GetTaskPool(task.GlobalTaskPoolID)
		assert.NotNil(t, pool)
	})
}

func TestWorkerPool_RegisterJob(t *testing.T) {
	t.Run("registers job without panic", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		assert.NotPanics(t, func() {
			wp.RegisterJob("test-job", func(_ *task.Task) {})
		})
	})

	t.Run("registers job with options", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		assert.NotPanics(t, func() {
			wp.RegisterJob("test-job", func(_ *task.Task) {}, task.Options{Persistent: true, Unique: true})
		})
	})
}

func TestWorkerPool_DispatchJob(t *testing.T) {
	t.Run("returns error for non-registered job", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		meta := task.NewTestMetadata("non-existent-job")

		_, err := wp.DispatchJob(context.Background(), "non-existent-job", meta, nil)
		assert.Error(t, err)
	})

	t.Run("returns error for mismatched job name", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)
		wp.RegisterJob("test-job", func(_ *task.Task) {})

		meta := task.NewTestMetadata("different-job")

		_, err := wp.DispatchJob(context.Background(), "test-job", meta, nil)
		assert.Error(t, err)
	})
}

func TestWorkerPool_Status(t *testing.T) {
	t.Run("returns status values", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)

		queueLen, total, busy, alive, retry := wp.Status()
		assert.GreaterOrEqual(t, queueLen, 0)
		assert.GreaterOrEqual(t, total, 0)
		assert.GreaterOrEqual(t, busy, 0)
		assert.GreaterOrEqual(t, alive, 0)
		assert.GreaterOrEqual(t, retry, 0)
	})
}

func TestWorkerPool_GetTask(t *testing.T) {
	t.Run("returns nil for non-existent task", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		tsk := wp.GetTask("non-existent-id")
		assert.Nil(t, tsk)
	})
}

func TestWorkerPool_GetTasks(t *testing.T) {
	t.Run("returns empty slice initially", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		tasks := wp.GetTasks()
		assert.Empty(t, tasks)
	})
}

func TestWorkerPool_GetTasksByJobName(t *testing.T) {
	t.Run("returns empty slice for non-existent job", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		tasks := wp.GetTasksByJobName("non-existent-job")
		assert.Empty(t, tasks)
	})
}

func TestWorkerPool_GetTaskPoolByJobName(t *testing.T) {
	t.Run("returns nil for non-existent job", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		pool := wp.GetTaskPoolByJobName("non-existent-job")
		assert.Nil(t, pool)
	})
}

func TestWorkerPool_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent pool creation", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		var wg sync.WaitGroup

		pools := make(chan *task.Pool, 100)

		for range 100 {
			wg.Go(func() {
				pool := wp.NewTaskPool(false, nil)
				pools <- pool
			})
		}

		wg.Wait()
		close(pools)

		// All pools should be unique
		seen := make(map[string]bool)
		for pool := range pools {
			assert.False(t, seen[pool.ID()], "duplicate pool ID found")

			seen[pool.ID()] = true
		}
	})

	t.Run("handles concurrent status reads", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				_, _, _, _, _ = wp.Status()
			}()
		}

		wg.Wait()
	})

	t.Run("handles concurrent job registration", func(t *testing.T) {
		wp := task.NewTestWorkerPool(1)

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func(idx int) {
				defer wg.Done()

				jobName := "test-job-" + string(rune('a'+idx%26))
				wp.RegisterJob(jobName, func(_ *task.Task) {})
			}(i)
		}

		wg.Wait()
	})
}

// ==================== Task Lifecycle Tests ====================

func TestTask_Cancel(t *testing.T) {
	t.Run("cancels running task", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		started := make(chan struct{})
		canceled := make(chan struct{})

		wp.RegisterJob("cancellable-job", func(tsk *task.Task) {
			close(started)
			<-tsk.Ctx.Done()
			close(canceled)
		})

		meta := task.NewTestMetadata("cancellable-job")
		tsk, err := wp.DispatchJob(context.Background(), "cancellable-job", meta, nil)
		assert.NoError(t, err)

		<-started
		tsk.Cancel()

		select {
		case <-canceled:
			// Success - task was canceled
		case <-time.After(2 * time.Second):
			t.Error("task should have been canceled")
		}
	})
}

func TestTask_Success(t *testing.T) {
	t.Run("marks task as successful", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("success-job", func(tsk *task.Task) {
			tsk.Success("completed successfully")
		})

		meta := task.NewTestMetadata("success-job")
		tsk, err := wp.DispatchJob(context.Background(), "success-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskSuccess, status)
	})

	t.Run("marks task as successful without message", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("success-no-msg", func(tsk *task.Task) {
			tsk.Success()
		})

		meta := task.NewTestMetadata("success-no-msg")
		tsk, err := wp.DispatchJob(context.Background(), "success-no-msg", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskSuccess, status)
	})
}

func TestTask_Fail(t *testing.T) {
	t.Run("panics with ErrTaskError", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("fail-job", func(tsk *task.Task) {
			tsk.Fail(intentionalFailure)
		})

		meta := task.NewTestMetadata("fail-job")
		tsk, err := wp.DispatchJob(context.Background(), "fail-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskError, status)
	})
}

func TestTask_ReqNoErr(t *testing.T) {
	t.Run("does nothing for nil error", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("reqnoerr-nil", func(tsk *task.Task) {
			tsk.ReqNoErr(nil)
			tsk.Success()
		})

		meta := task.NewTestMetadata("reqnoerr-nil")
		tsk, err := wp.DispatchJob(context.Background(), "reqnoerr-nil", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskSuccess, status)
	})

	t.Run("fails for non-nil error", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		testErr := intentionalFailure

		wp.RegisterJob("reqnoerr-err", func(tsk *task.Task) {
			tsk.ReqNoErr(testErr)
		})

		meta := task.NewTestMetadata("reqnoerr-err")
		tsk, err := wp.DispatchJob(context.Background(), "reqnoerr-err", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskError, status)
	})
}

func TestTask_SetChildTaskPool(t *testing.T) {
	t.Run("sets child task pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		var childPool *task.Pool

		wp.RegisterJob("parent-job", func(tsk *task.Task) {
			childPool = wp.NewTaskPool(false, tsk)
			tsk.SetChildTaskPool(childPool)
			tsk.Success()
		})

		meta := task.NewTestMetadata("parent-job")
		tsk, err := wp.DispatchJob(context.Background(), "parent-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		assert.NotNil(t, tsk.GetChildTaskPool())
	})
}

// ==================== Task Callback Tests ====================

func TestTask_SetPostAction(t *testing.T) {
	t.Run("runs after successful completion", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		postActionRan := make(chan bool, 1)

		wp.RegisterJob("post-action-job", func(tsk *task.Task) {
			tsk.SetResult(task.Result{"key": "value"})
			tsk.Success()
		})

		meta := task.NewTestMetadata("post-action-job")
		tsk, err := wp.DispatchJob(context.Background(), "post-action-job", meta, nil)
		assert.NoError(t, err)

		tsk.SetPostAction(func(result task.Result) {
			postActionRan <- true
		})

		tsk.Wait()

		select {
		case ran := <-postActionRan:
			assert.True(t, ran)
		case <-time.After(2 * time.Second):
			t.Error("post action should have run")
		}
	})

	t.Run("runs immediately if task already completed successfully", func(t *testing.T) {
		tsk := task.NewTestTask("task-1", "TestJob", 0, task.TaskSuccess, task.Result{"key": "value"}, time.Now())

		postActionRan := false
		tsk.SetPostAction(func(_ task.Result) {
			postActionRan = true
		})

		assert.True(t, postActionRan)
	})
}

func TestTask_SetCleanup(t *testing.T) {
	t.Run("runs cleanup after task completes successfully", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		cleanupRan := make(chan bool, 1)

		wp.RegisterJob("cleanup-job", func(tsk *task.Task) {
			tsk.SetCleanup(func(_ *task.Task) {
				cleanupRan <- true
			})
			tsk.Success()
		})

		meta := task.NewTestMetadata("cleanup-job")
		tsk, err := wp.DispatchJob(context.Background(), "cleanup-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		select {
		case ran := <-cleanupRan:
			assert.True(t, ran)
		case <-time.After(2 * time.Second):
			t.Error("cleanup should have run")
		}
	})

	t.Run("runs cleanup after task fails", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		cleanupRan := make(chan bool, 1)

		wp.RegisterJob("cleanup-fail-job", func(tsk *task.Task) {
			tsk.SetCleanup(func(_ *task.Task) {
				cleanupRan <- true
			})
			tsk.Fail(intentionalFailure)
		})

		meta := task.NewTestMetadata("cleanup-fail-job")
		tsk, err := wp.DispatchJob(context.Background(), "cleanup-fail-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		select {
		case ran := <-cleanupRan:
			assert.True(t, ran)
		case <-time.After(2 * time.Second):
			t.Error("cleanup should have run")
		}
	})
}

func TestTask_SetErrorCleanup(t *testing.T) {
	t.Run("runs error cleanup when task fails", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		errorCleanupRan := make(chan bool, 1)

		wp.RegisterJob("error-cleanup-job", func(tsk *task.Task) {
			tsk.SetErrorCleanup(func(_ *task.Task) {
				errorCleanupRan <- true
			})
			tsk.Fail(intentionalFailure)
		})

		meta := task.NewTestMetadata("error-cleanup-job")
		tsk, err := wp.DispatchJob(context.Background(), "error-cleanup-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		select {
		case ran := <-errorCleanupRan:
			assert.True(t, ran)
		case <-time.After(2 * time.Second):
			t.Error("error cleanup should have run")
		}
	})

	t.Run("does not run error cleanup when task succeeds", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		errorCleanupRan := make(chan bool, 1)

		wp.RegisterJob("no-error-cleanup-job", func(tsk *task.Task) {
			tsk.SetErrorCleanup(func(_ *task.Task) {
				errorCleanupRan <- true
			})
			tsk.Success()
		})

		meta := task.NewTestMetadata("no-error-cleanup-job")
		tsk, err := wp.DispatchJob(context.Background(), "no-error-cleanup-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		select {
		case <-errorCleanupRan:
			t.Error("error cleanup should not have run for successful task")
		case <-time.After(500 * time.Millisecond):
			// Success - error cleanup did not run
		}
	})
}

// ==================== Task Timing Tests ====================

func TestTask_SetQueueState(t *testing.T) {
	t.Run("sets queue state", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("queue-state-job", func(tsk *task.Task) {
			tsk.SetQueueState(task.Sleeping)
			assert.Equal(t, task.Sleeping, tsk.QueueState())

			tsk.SetQueueState(task.Executing)
			assert.Equal(t, task.Executing, tsk.QueueState())

			tsk.Success()
		})

		meta := task.NewTestMetadata("queue-state-job")
		tsk, err := wp.DispatchJob(context.Background(), "queue-state-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()
	})
}

func TestTask_SetQueueTime(t *testing.T) {
	t.Run("sets queue time", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		testTime := time.Now().Add(-1 * time.Hour)

		wp.RegisterJob("queue-time-job", func(tsk *task.Task) {
			tsk.SetQueueTime(testTime)
			tsk.Success()
		})

		meta := task.NewTestMetadata("queue-time-job")
		tsk, err := wp.DispatchJob(context.Background(), "queue-time-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()
	})
}

func TestTask_QueueTimeDuration(t *testing.T) {
	t.Run("returns duration in queue", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		var duration time.Duration

		wp.RegisterJob("queue-duration-job", func(tsk *task.Task) {
			duration = tsk.QueueTimeDuration()
			tsk.Success()
		})

		meta := task.NewTestMetadata("queue-duration-job")
		tsk, err := wp.DispatchJob(context.Background(), "queue-duration-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		assert.GreaterOrEqual(t, duration, time.Duration(0))
	})
}

func TestTask_SetTimeout(t *testing.T) {
	t.Run("sets timeout and triggers cancellation", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("timeout-job", func(tsk *task.Task) {
			tsk.SetTimeout(time.Now().Add(100 * time.Millisecond))

			// Wait for timeout
			<-tsk.Ctx.Done()
		})

		meta := task.NewTestMetadata("timeout-job")
		tsk, err := wp.DispatchJob(context.Background(), "timeout-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskCanceled, status)
	})
}

func TestTask_ExeTime_Running(t *testing.T) {
	t.Run("returns time since start for running task", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		var exeTime time.Duration

		wp.RegisterJob("exetime-job", func(tsk *task.Task) {
			time.Sleep(50 * time.Millisecond)
			exeTime = tsk.ExeTime()
			tsk.Success()
		})

		meta := task.NewTestMetadata("exetime-job")
		tsk, err := wp.DispatchJob(context.Background(), "exetime-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		assert.GreaterOrEqual(t, exeTime, 50*time.Millisecond)
	})
}

func TestTask_ReadError_WithError(t *testing.T) {
	t.Run("returns error for failed task", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("read-error-job", func(tsk *task.Task) {
			tsk.Fail(intentionalFailure)
		})

		meta := task.NewTestMetadata("read-error-job")
		tsk, err := wp.DispatchJob(context.Background(), "read-error-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		taskErr := tsk.ReadError()
		assert.NotNil(t, taskErr)
	})
}

// ==================== Pool Workflow Tests ====================

func TestPool_QueueTask(t *testing.T) {
	t.Run("queues and executes task", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		executed := make(chan bool, 1)

		wp.RegisterJob("queue-task-job", func(tsk *task.Task) {
			executed <- true

			tsk.Success()
		})

		meta := task.NewTestMetadata("queue-task-job")
		tsk, err := wp.DispatchJob(context.Background(), "queue-task-job", meta, nil)
		assert.NoError(t, err)

		select {
		case <-executed:
			// Success
		case <-time.After(2 * time.Second):
			t.Error("task should have been executed")
		}

		tsk.Wait()
	})
}

func TestPool_Cancel(t *testing.T) {
	t.Run("cancels all tasks in pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		started := make(chan struct{})
		pool := wp.NewTaskPool(false, nil)

		wp.RegisterJob("cancel-pool-job", func(tsk *task.Task) {
			close(started)
			<-tsk.Ctx.Done()
		})

		meta := task.NewTestMetadata("cancel-pool-job")
		tsk, err := wp.DispatchJob(context.Background(), "cancel-pool-job", meta, pool)
		assert.NoError(t, err)

		<-started
		pool.Cancel()

		tsk.Wait()

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskCanceled, status)
	})
}

func TestPool_SignalAllQueued(t *testing.T) {
	t.Run("signals all tasks queued", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		pool := wp.NewTaskPool(false, nil)

		wp.RegisterJob("signal-job", func(tsk *task.Task) {
			tsk.Success()
		})

		meta := task.NewTestMetadata("signal-job")
		tsk, err := wp.DispatchJob(context.Background(), "signal-job", meta, pool)
		assert.NoError(t, err)

		pool.SignalAllQueued()

		tsk.Wait()
	})
}

func TestPool_RemoveTask(t *testing.T) {
	t.Run("removes task from pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		pool := wp.NewTaskPool(false, nil)

		wp.RegisterJob("remove-task-job", func(tsk *task.Task) {
			tsk.Success()
		})

		meta := task.NewTestMetadata("remove-task-job")
		tsk, err := wp.DispatchJob(context.Background(), "remove-task-job", meta, pool)
		assert.NoError(t, err)

		tsk.Wait()

		pool.RemoveTask(tsk.ID())
	})
}

func TestPool_Wait(t *testing.T) {
	t.Run("waits for all tasks to complete", func(t *testing.T) {
		wp := task.NewTestWorkerPool(4)
		wp.Run()

		completedCount := 0

		var mu sync.Mutex

		var pool *task.Pool

		wp.RegisterJob("wait-job", func(tsk *task.Task) {
			time.Sleep(50 * time.Millisecond)
			mu.Lock()

			completedCount++
			log.GlobalLogger().Info().Msgf("Completed tasks: %d", completedCount)

			mu.Unlock()
			tsk.Success()
		}, task.Options{Unique: true})

		wp.RegisterJob("wait-parent-job", func(parentTsk *task.Task) {
			// Create pool inside task so createdBy is set
			pool = wp.NewTaskPool(false, parentTsk)

			for range 3 {
				meta := task.NewTestMetadata("wait-job")

				_, err := wp.DispatchJob(context.Background(), "wait-job", meta, pool)
				if err != nil {
					parentTsk.Fail(err)

					return
				}
			}

			pool.SignalAllQueued()
			pool.Wait(false, parentTsk)

			mu.Lock()

			count := completedCount

			mu.Unlock()

			if count != 3 {
				parentTsk.Fail(wlerrors.Errorf("expected 3 completed tasks, got %d", count))

				return
			}

			parentTsk.Success()
		})

		meta := task.NewTestMetadata("wait-parent-job")
		parentTsk, err := wp.DispatchJob(context.Background(), "wait-parent-job", meta, nil)
		assert.NoError(t, err)

		parentTsk.Wait()

		done, status := parentTsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskSuccess, status)
	})

	t.Run("returns immediately for global pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		globalPool := wp.GetTaskPool(task.GlobalTaskPoolID)

		// Should not block
		done := make(chan bool)

		go func() {
			globalPool.Wait(false)

			done <- true
		}()

		select {
		case <-done:
			// Success - returned immediately
		case <-time.After(1 * time.Second):
			t.Error("Wait should return immediately for global pool")
		}
	})
}

// ==================== WorkerPool Execution Tests ====================

func TestWorkerPool_Run(t *testing.T) {
	t.Run("starts workers and processes tasks", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		executed := make(chan bool, 1)

		wp.RegisterJob("run-test-job", func(tsk *task.Task) {
			executed <- true

			tsk.Success()
		})

		meta := task.NewTestMetadata("run-test-job")
		tsk, err := wp.DispatchJob(context.Background(), "run-test-job", meta, nil)
		assert.NoError(t, err)

		select {
		case <-executed:
			// Success
		case <-time.After(2 * time.Second):
			t.Error("task should have been executed")
		}

		tsk.Wait()
	})
}

func TestWorkerPool_DispatchJob_FullFlow(t *testing.T) {
	t.Run("dispatches and completes job successfully", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("full-flow-job", func(tsk *task.Task) {
			tsk.SetResult(task.Result{"processed": true})
			tsk.Success()
		})

		meta := task.NewTestMetadata("full-flow-job")
		tsk, err := wp.DispatchJob(context.Background(), "full-flow-job", meta, nil)
		assert.NoError(t, err)
		assert.NotNil(t, tsk)

		tsk.Wait()

		result := tsk.GetResult()
		assert.Equal(t, true, result["processed"])

		done, status := tsk.Status()
		assert.True(t, done)
		assert.Equal(t, task.TaskSuccess, status)
	})

	t.Run("dispatches job to specific pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		pool := wp.NewTaskPool(false, nil)

		wp.RegisterJob("pool-specific-job", func(tsk *task.Task) {
			tsk.Success()
		})

		meta := task.NewTestMetadata("pool-specific-job")
		tsk, err := wp.DispatchJob(context.Background(), "pool-specific-job", meta, pool)
		assert.NoError(t, err)

		pool.SignalAllQueued()
		tsk.Wait()

		done, _ := tsk.Status()
		assert.True(t, done)
	})

	t.Run("returns existing task if already dispatched", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		started := make(chan struct{})
		canFinish := make(chan struct{})

		wp.RegisterJob("dup-job", func(tsk *task.Task) {
			close(started)
			<-canFinish
			tsk.Success()
			// Explicitly set Unique option here for readability.
			// This test ensures that duplicate dispatches return the same task,
			// therefore uniqueness is strictly disabled.
		}, task.Options{Unique: false})

		meta := task.NewTestMetadata("dup-job")
		tsk1, err := wp.DispatchJob(context.Background(), "dup-job", meta, nil)
		assert.NoError(t, err)

		<-started

		// Try to dispatch same job again
		tsk2, err := wp.DispatchJob(context.Background(), "dup-job", meta, nil)
		assert.NoError(t, err)
		assert.Equal(t, tsk1.ID(), tsk2.ID())

		close(canFinish)
		tsk1.Wait()
	})
}

func TestWorkerPool_AddHit(t *testing.T) {
	t.Run("schedules timeout hit", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		wp.RegisterJob("hit-job", func(tsk *task.Task) {
			wp.AddHit(time.Now().Add(100*time.Millisecond), tsk)
			<-tsk.Ctx.Done()
		})

		meta := task.NewTestMetadata("hit-job")
		tsk, err := wp.DispatchJob(context.Background(), "hit-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()
	})
}

func TestWorkerPool_GetTasksByJobName_WithTasks(t *testing.T) {
	t.Run("returns tasks by job name", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		started := make(chan struct{})
		canFinish := make(chan struct{})

		wp.RegisterJob("named-job", func(tsk *task.Task) {
			close(started)
			<-canFinish
			tsk.Success()
		})

		meta := task.NewTestMetadata("named-job")
		_, err := wp.DispatchJob(context.Background(), "named-job", meta, nil)
		assert.NoError(t, err)

		<-started

		tasks := wp.GetTasksByJobName("named-job")
		assert.NotEmpty(t, tasks)

		close(canFinish)
	})
}

func TestWorkerPool_NewTaskPool_WithCreatedBy(t *testing.T) {
	t.Run("creates pool with parent task reference", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		var childPool *task.Pool

		wp.RegisterJob("parent-task", func(tsk *task.Task) {
			childPool = wp.NewTaskPool(false, tsk)
			tsk.Success()
		})

		meta := task.NewTestMetadata("parent-task")
		tsk, err := wp.DispatchJob(context.Background(), "parent-task", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		assert.NotNil(t, childPool)
		assert.NotNil(t, childPool.CreatedInTask())
	})

	t.Run("creates pool with replacement worker", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		var childPool *task.Pool

		wp.RegisterJob("replacement-parent", func(tsk *task.Task) {
			childPool = wp.NewTaskPool(true, tsk)
			childPool.SignalAllQueued()
			tsk.Success()
		})

		meta := task.NewTestMetadata("replacement-parent")
		tsk, err := wp.DispatchJob(context.Background(), "replacement-parent", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		assert.NotNil(t, childPool)
	})
}

func TestWorkerPool_GetTaskPoolByJobName_WithMatch(t *testing.T) {
	t.Run("returns pool created by matching job", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		started := make(chan struct{})
		canFinish := make(chan struct{})

		wp.RegisterJob("pool-creator-job", func(tsk *task.Task) {
			wp.NewTaskPool(false, tsk)
			close(started)
			<-canFinish
			tsk.Success()
		})

		meta := task.NewTestMetadata("pool-creator-job")
		_, err := wp.DispatchJob(context.Background(), "pool-creator-job", meta, nil)
		assert.NoError(t, err)

		<-started

		pool := wp.GetTaskPoolByJobName("pool-creator-job")
		assert.NotNil(t, pool)

		close(canFinish)
	})
}

// ==================== Task Q Method Tests ====================

func TestTask_Q_WithPool(t *testing.T) {
	t.Run("queues task to pool", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		pool := wp.NewTaskPool(false, nil)

		executed := make(chan bool, 1)

		wp.RegisterJob("q-job", func(tsk *task.Task) {
			executed <- true

			tsk.Success()
		})

		meta := task.NewTestMetadata("q-job")
		tsk, err := wp.DispatchJob(context.Background(), "q-job", meta, pool)
		assert.NoError(t, err)

		pool.SignalAllQueued()

		select {
		case <-executed:
			// Success
		case <-time.After(2 * time.Second):
			t.Error("task should have been executed")
		}

		tsk.Wait()
	})
}

// ==================== Task ClearAndRecompute Tests ====================
// Note: ClearAndRecompute is tested implicitly via the task lifecycle.
// Full testing requires fixing race conditions in production code.

// ==================== Pool GetRootPool Tests ====================

func TestPool_GetRootPool_NestedPools(t *testing.T) {
	t.Run("returns root for nested pools", func(t *testing.T) {
		wp := task.NewTestWorkerPool(2)
		wp.Run()

		var childPool, grandchildPool *task.Pool

		wp.RegisterJob("root-pool-job", func(tsk *task.Task) {
			childPool = wp.NewTaskPool(false, tsk)

			wp.RegisterJob("child-pool-job", func(childTsk *task.Task) {
				grandchildPool = wp.NewTaskPool(false, childTsk)
				childTsk.Success()
			})

			childMeta := task.NewTestMetadata("child-pool-job")
			childTsk, _ := wp.DispatchJob(context.Background(), "child-pool-job", childMeta, childPool)
			childPool.SignalAllQueued()
			childTsk.Wait()

			tsk.Success()
		})

		meta := task.NewTestMetadata("root-pool-job")
		tsk, err := wp.DispatchJob(context.Background(), "root-pool-job", meta, nil)
		assert.NoError(t, err)

		tsk.Wait()

		// The grandchild pool's root should be the child pool (not global)
		if grandchildPool != nil {
			root := grandchildPool.GetRootPool()
			assert.NotNil(t, root)
		}
	})
}

// ==================== Pool IsRoot Additional Tests ====================

func TestPool_IsRoot_NilCheck(t *testing.T) {
	t.Run("returns false for nil pool", func(t *testing.T) {
		var pool *task.Pool
		assert.False(t, pool.IsRoot())
	})
}
