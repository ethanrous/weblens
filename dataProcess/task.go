package dataProcess

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/websocket"
)

type task struct {
	taskId        types.TaskId
	taskPool      *taskPool
	childTaskPool *taskPool
	caster websocket.BroadcasterAgent
	work          taskHandler
	taskType      types.TaskType
	metadata      types.TaskMetadata
	result        types.TaskResult
	persistent    bool
	queueState    taskQueueState

	updateMu *sync.RWMutex

	err any

	timeout   time.Time
	timerLock sync.Mutex

	exitStatus types.TaskExitStatus // "success", "error" or "cancelled"

	postAction func(result types.TaskResult)

	// Function to be run to clean up when the task completes, no matter the exit status
	cleanup func()

	// Function to be run to clean up if the task errors
	errorCleanup func()

	sw internal.Stopwatch

	// signal is used for signaling a task to change behavior after it has been queued,
	// to exit prematurely, for example. The signalChan serves the same purpose, but is
	// used when a task might block waiting for another channel.
	// Key: 1 is exit,
	signal atomic.Int64
	signalChan chan int

	waitMu sync.Mutex
}

type taskQueueState string

const (
	PreQueued taskQueueState = "pre-queued"
	InQueue   taskQueueState = "in-queue"
	Executing taskQueueState = "executing"
	Exited    taskQueueState = "exited"
)

func (t *task) TaskId() types.TaskId {
	return t.taskId
}

func (t *task) TaskType() types.TaskType {
	return t.taskType
}

func (t *task) GetTaskPool() types.TaskPool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	return t.taskPool
}

func (t *task) setTaskPoolInternal(pool *taskPool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.taskPool = pool
}

func (t *task) GetChildTaskPool() types.TaskPool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	return t.childTaskPool
}

func (t *task) setChildTaskPool(pool *taskPool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.childTaskPool = pool
}

// Status returns a boolean representing if a task has completed, and a string describing its exit type, if completed.
func (t *task) Status() (bool, types.TaskExitStatus) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	return t.queueState == Exited, t.exitStatus
}

func (t *task) SetCaster(c websocket.BroadcasterAgent) {
	t.caster = c
}

// Q queues task on given taskPool tp,
// if tp is nil, will default to the global task pool.
// Essentially an alias for tp.QueueTask(t), so you can
// NewTask(...).Q(). Returns the given task to further support this
func (t *task) Q(tp types.TaskPool) types.Task {
	if tp == nil {
		wlog.Error.Println("nil task pool")
		return nil
		// tp = GetGlobalQueue()
	}
	err := tp.QueueTask(t)
	if err != nil {
		wlog.ShowErr(err)
		return nil
	}

	return t
}

// Wait Block until a task is finished. "Finished" can define success, failure, or cancel
func (t *task) Wait() types.Task {
	if t == nil {
		return t
	}
	t.updateMu.Lock()
	if t.queueState == Exited {
		return t
	}
	t.updateMu.Unlock()
	t.waitMu.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	t.waitMu.Unlock()

	return t
}

// Cancel Unknowable if this is the last operation of a task, so t.success()
// will not have an effect after a task is cancelled. t.error() may
// override the exit status in special cases, such as a timeout,
// which is both an error and a reason for cancellation.
//
// Cancellations are always external to the task. From within the
// body of the task, either error or success should be called.
// If a task finds itself not required to continue, success should
// be returned
func (t *task) Cancel() {
	if t == nil {
		return
	}

	if t.queueState == Exited || t.signal.Load() != 0 {
		return
	}
	t.signal.Store(1)
	t.signalChan <- 1
	if t.exitStatus == TaskNoStatus {
		t.exitStatus = TaskCanceled
	}

	// Do not exit task here, so that .Wait() -ing on a task will wait until the task actually exits,
	// before starting again
	// t.queueState = Exited
}

// CheckExit should be used intermittently to check if the task should exit.
// If the task should exit, it panics back to the top of safety work
func (t *task) CheckExit() {
	if t.signal.Load() != 0 {
		panic(ErrTaskExit)
	}
}

func (t *task) ClearAndRecompute() {
	t.Cancel()
	t.Wait()

	t.updateMu.Lock()
	t.exitStatus = TaskNoStatus
	t.signal.Store(0)
	t.queueState = PreQueued
	t.updateMu.Unlock()

	t.waitMu.TryLock()

	for k := range t.result {
		delete(t.result, k)
	}

	if t.err != nil {
		wlog.Warning.Printf("Retrying task (%s) that has previous error: %v", t.TaskId(), t.err)
		t.err = nil
	}

	err := t.taskPool.QueueTask(t)
	if err != nil {
		return
	}

	t.taskPool.workerPool.taskMap[t.taskId] = t
}

func (t *task) GetResult(resultKey string) any {
	if t.result == nil {
		return nil
	}

	return t.result[resultKey]
}

func (t *task) GetResults() types.TaskResult {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	if t.result == nil {
		t.result = types.TaskResult{}
	}
	return t.result
}

func (t *task) GetMeta() types.TaskMetadata {
	return t.metadata
}

// Set the error of the task. This *should* be the last operation performed before returning from the task.
// However, sometimes that is not possible, so we must check if the task has been cancelled before setting
// the error, as errors occurring inside the task body, after a task is cancelled, are not valid.
// If an error has caused the task to be cancelled, t.Cancel() must be called after t.error()
func (t *task) error(err error) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	// Run the cleanup routine for errors, if any
	if t.errorCleanup != nil {
		t.errorCleanup()
		t.errorCleanup = nil
	}

	// If we have already called cancel, do not set any error
	// E.g. A file is being moved, so we cancel all tasks on it,
	// and move it in the filesystem. The task goes to find the file, it can't (because it was moved)
	// and throws this error. Now we are here and we realize the task was canceled, so that error is not valid
	if t.queueState == Exited {
		return
	}

	t.err = err
	t.queueState = Exited
	t.exitStatus = TaskError

	t.sw.Lap("Task exited due to error")
	t.sw.Stop()

}

func (t *task) ErrorAndExit(err error, info ...any) {
	if err == nil {
		return
	}

	wlog.ShowErr(err, fmt.Sprintf("Task %s exited with an error", t.TaskId()))
	// _, filename, line, _ := runtime.Caller(1)
	// util.ErrorCatcher.Printf("Task %s exited with an error\n\t%s:%d %s\n", t.TaskId(), filename, line, err.Error())
	if len(info) != 0 {
		wlog.ErrorCatcher.Printf("Reported by task: %s", fmt.Sprint(info...))
	}
	t.error(err)
	panic(ErrTaskError)
}

// SetPostAction takes a function to be run after the task has successfully completed
// with the task results as the input of the function
func (t *task) SetPostAction(action func(types.TaskResult)) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.postAction = action

	// If the task has already completed, run the post task in this thread instead
	if t.exitStatus == TaskSuccess {
		t.postAction(t.result)
	}
}

// Pass a function to run if the task throws an error, in theory
// to cleanup any half-processed state that could litter if not finished
func (t *task) SetErrorCleanup(cleanup func()) {
	t.errorCleanup = cleanup
}

func (t *task) SetCleanup(cleanup func()) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.cleanup = cleanup
}

func (t *task) ReadError() any {
	return t.err
}

func (t *task) success(msg ...any) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	if t.cleanup != nil {
		t.cleanup()
		t.cleanup = nil
	}

	t.queueState = Exited
	t.exitStatus = TaskSuccess
	if len(msg) != 0 {
		wlog.Info.Println("Task succeeded with a message:", fmt.Sprint(msg...))
	}

	t.sw.Stop()
}

func (t *task) setTimeout(timeout time.Time) {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()
	t.timeout = timeout
	t.GetTaskPool().GetWorkerPool()
}

func (t *task) ClearTimeout() {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()
	t.timeout = time.Unix(0, 0)
}

func (t *task) setResult(results types.TaskResult) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.result = results
	// if t.result == nil {
	// 	t.result = make(map[string]string)
	// }

	// for _, pair := range fields {
	// 	t.result[pair.Key] = pair.Val
	// }
}

// Add a lap in the tasks stopwatch
func (t *task) SwLap(label string) {
	t.sw.Lap(label)
}

// Add a lap in the tasks stopwatch
func (t *task) ExeTime() time.Duration {
	return t.sw.GetTotalTime(true)
}
