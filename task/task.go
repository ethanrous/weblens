package task

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
)

type Id = string
type TaskExitStatus string
type TaskResultKey string
type TaskResult map[TaskResultKey]any

func (tr TaskResult) ToMap() map[string]any {
	m := map[string]any{}
	for k, v := range tr {
		m[string(k)] = v
	}
	return m
}

type Task struct {
	timeout  time.Time
	metadata TaskMetadata

	err error

	sw internal.Stopwatch

	taskPool      *TaskPool
	childTaskPool *TaskPool
	work          TaskHandler
	result        TaskResult

	// Function to be run to clean up when the task completes, only if the task is successful
	postAction func(result TaskResult)

	// Function to run whenever the task results update
	resultsCallback func(result TaskResult)

	signalChan chan int

	taskId     Id
	jobName    string
	queueState QueueState

	exitStatus TaskExitStatus // "success", "error" or "cancelled"

	// Function to be run to clean up when the task completes, no matter the exit status
	cleanup []TaskHandler

	// Function to be run to clean up if the task errors
	errorCleanup []TaskHandler

	// signal is used for signaling a task to change behavior after it has been queued,
	// to exit prematurely, for example. The signalChan serves the same purpose, but is
	// used when a task might block waiting for another channel.
	// Key: 1 is exit,
	signal atomic.Int64

	updateMu sync.RWMutex

	timerLock sync.Mutex

	waitMu     sync.Mutex
	persistent bool
}

type QueueState string

const (
	PreQueued QueueState = "pre-queued"
	InQueue   QueueState = "in-queue"
	Executing QueueState = "executing"
	Exited    QueueState = "exited"
)

func (t *Task) TaskId() Id {
	return t.taskId
}

func (t *Task) JobName() string {
	return t.jobName
}

func (t *Task) GetTaskPool() *TaskPool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	return t.taskPool
}

func (t *Task) setTaskPoolInternal(pool *TaskPool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.taskPool = pool
}

func (t *Task) GetChildTaskPool() *TaskPool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	return t.childTaskPool
}

func (t *Task) SetChildTaskPool(pool *TaskPool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.childTaskPool = pool
}

// Status returns a boolean representing if a task has completed, and a string describing its exit type, if completed.
func (t *Task) Status() (bool, TaskExitStatus) {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	return t.queueState == Exited, t.exitStatus
}

// Q queues task on given taskPool tp,
// if tp is nil, will default to the global task pool.
// Essentially an alias for tp.QueueTask(t), so you can
// NewTask(...).Q(). Returns the given task to further support this
func (t *Task) Q(tp *TaskPool) *Task {
	if tp == nil {
		log.Error.Println("nil task pool")
		return nil
		// tp = GetGlobalQueue()
	}
	err := tp.QueueTask(t)
	if err != nil {
		log.ShowErr(err)
		return nil
	}

	return t
}

// Wait Block until a task is finished. "Finished" can define success, failure, or cancel
func (t *Task) Wait() *Task {
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
func (t *Task) Cancel() {
	log.Trace.Printf("Cancelling task T[%s]", t.taskId)

	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	if t.queueState == Exited || t.signal.Load() != 0 {
		return
	}
	t.signal.Store(1)
	t.signalChan <- 1
	if t.exitStatus == TaskNoStatus {
		t.queueState = Exited
		t.exitStatus = TaskCanceled
	}
	if t.childTaskPool != nil {
		t.childTaskPool.Cancel()
	}

	// Do not exit task here, so that .Wait() -ing on a task will wait until the task actually exits,
	// before starting again
	// t.queueState = Exited
}

// ExitIfSignaled should be used intermittently to check if the task should exit.
// If the task should exit, it panics back to the top of safety work
func (t *Task) ExitIfSignaled() {
	if t.CheckExit() {
		panic(werror.ErrTaskExit)
	}
}

func (t *Task) CheckExit() bool {
	return t.signal.Load() != 0
}

func (t *Task) ClearAndRecompute() {
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
		log.Warning.Printf("Retrying task (%s) that has previous error: %v", t.TaskId(), t.err)
		t.err = nil
	}

	err := t.taskPool.QueueTask(t)
	if err != nil {
		return
	}

	t.taskPool.workerPool.taskMap[t.taskId] = t
}

func (t *Task) GetResult(resultKey TaskResultKey) any {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	if t.result == nil {
		return nil
	}

	if resultKey == "" {
		return maps.Clone(t.result)
	}

	return t.result[resultKey]
}

func (t *Task) GetResults() TaskResult {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	if t.result == nil {
		t.result = TaskResult{}
	}
	return maps.Clone(t.result)
}

func (t *Task) GetMeta() TaskMetadata {
	return t.metadata
}

/*
Manipulate is used to change the metadata of a task while it is running.
This can be useful to have a task be waiting for input from a client,
and this function can be used to send that data to the task via a chan, for example.
*/
func (t *Task) Manipulate(fn func(meta TaskMetadata) error) error {
	return fn(t.metadata)
}

// Set the error of the task. This *should* be the last operation performed before returning from the task.
// However, sometimes that is not possible, so we must check if the task has been cancelled before setting
// the error, as errors occurring inside the task body, after a task is cancelled, are not valid.
// If an error has caused the task to be cancelled, t.Cancel() must be called after t.error()
func (t *Task) error(err error) {
	t.updateMu.Lock()

	if t.queueState != Exited {
		log.Trace.Println("Setting task Error")
		t.err = err
		t.queueState = Exited
		t.exitStatus = TaskError
	}

	// If we have already called cancel, do not set any error
	// E.g. A file is being moved, so we cancel all tasks on it,
	// and move it in the filesystem. The task goes to find the file, it can't (because it was moved)
	// and throws this error. Now we are here and we realize the task was canceled, so that error is not valid
	if t.queueState == Exited {
		t.updateMu.Unlock()
		return
	}

	t.updateMu.Unlock()
	t.sw.Lap("Task exited due to error")
	t.sw.Stop()
}

// ReqNoErr is a wrapper around t.Fail, but only fails if the error is not nil
func (t *Task) ReqNoErr(err error) {
	if err == nil {
		return
	}

	t.Fail(err)
}

// Fail will set the error on the task, and then panic with ErrTaskError, which informs the worker recovery
// function to exit the task with the error that is set, and not treat it as a real panic.
func (t *Task) Fail(err error) {
	if err == nil {
		panic(werror.Errorf("Trying to fail task with nil error"))
	}

	t.error(err)
	panic(werror.ErrTaskError)
}

func (t *Task) GetSignalChan() chan int {
	return t.signalChan
}

// SetPostAction takes a function to be run after the task has successfully completed
// with the task results as the input of the function
func (t *Task) SetPostAction(action func(TaskResult)) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	// If the task has already completed, run the post task in this thread instead
	if t.exitStatus == TaskSuccess {
		t.postAction(t.result)
	} else if t.exitStatus == TaskNoStatus {
		t.postAction = action
	}
}

// SetErrorCleanup works the same as t.SetCleanup(), but only runs if the task errors
func (t *Task) SetErrorCleanup(cleanup TaskHandler) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.errorCleanup = append(t.errorCleanup, cleanup)
}

// SetCleanup takes a function to be run after the task has completed, no matter the exit status.
// Many cleanup functions can be registered to run in sequence after the task completes. The cleanup
// functions are run in the order they are registered.
// Modifications to the task state should not be made in the cleanup functions (i.e. read-only), as the task has already completed, and may result in a deadlock.
// If the task has already completed, this function will NOT be called. Therefore, it is only safe to call SetCleanup() from inside of a task handler.
// If you want to register a function from outside the task handler, or to run after the task has completed successfully, use t.SetPostAction() instead.
func (t *Task) SetCleanup(cleanup TaskHandler) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.cleanup = append(t.cleanup, cleanup)
}

func (t *Task) ReadError() error {
	switch t.err.(type) {
	case nil:
		return nil
	case error:
		return t.err
	default:
		return werror.Errorf("%s", t.err)
	}
}

func (t *Task) Success(msg ...any) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.queueState = Exited
	t.exitStatus = TaskSuccess
	if len(msg) != 0 {
		log.Info.Println("Task succeeded with a message:", fmt.Sprint(msg...))
	}

	t.sw.Stop()
}

func (t *Task) SetTimeout(timeout time.Time) {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()
	t.timeout = timeout
	t.GetTaskPool().GetWorkerPool()
}

func (t *Task) ClearTimeout() {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()
	t.timeout = time.Unix(0, 0)
}

// OnResult takes a function to be run when the task result changes
func (t *Task) OnResult(callback func(TaskResult)) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.resultsCallback = callback
}

func (t *Task) SetResult(results TaskResult) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	if t.result == nil {
		t.result = results
	} else {
		for k, v := range results {
			t.result[k] = v
		}
	}

	if t.resultsCallback != nil {
		t.resultsCallback(maps.Clone(t.result))
	}
}

// Add a lap in the tasks stopwatch
func (t *Task) SwLap(label string) {
	t.sw.Lap(label)
}

func (t *Task) ExeTime() time.Duration {
	return t.sw.GetTotalTime(true)
}

func globbyHash(charLimit int, dataToHash ...any) string {
	h := sha256.New()

	s := fmt.Sprint(dataToHash...)
	h.Write([]byte(s))

	return base64.URLEncoding.EncodeToString(h.Sum(nil))[:charLimit]
}

type TaskInterface interface {
	TaskId() Id
	JobName() string
	GetTaskPool() *TaskPool
	GetChildTaskPool() *TaskPool
	Status() (bool, TaskExitStatus)
	GetMeta() TaskMetadata
	GetResult(string) any
	GetResults() TaskResult

	Q(pool *TaskPool) *Task

	Wait() *Task
	Cancel()

	SwLap(string)
	ClearTimeout()

	ReadError() any
	ClearAndRecompute()
	SetPostAction(action func(TaskResult))
	SetCleanup(cleanup func())
	SetErrorCleanup(cleanup func())

	ExeTime() time.Duration
}

type TaskMetadata interface {
	JobName() string
	MetaString() string
	FormatToResult() TaskResult
	Verify() error
}

const (
	TaskSuccess  TaskExitStatus = "success"
	TaskCanceled TaskExitStatus = "cancelled"
	TaskError    TaskExitStatus = "error"
	TaskNoStatus TaskExitStatus = ""
)
