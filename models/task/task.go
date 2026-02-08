package task

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/rs/zerolog"
)

// Task represents a unit of work that can be queued and executed by a worker pool.
type Task struct {
	QueueTime  time.Time
	StartTime  time.Time
	FinishTime time.Time

	Ctx        context.Context
	ctxMu      sync.RWMutex
	cancelFunc context.CancelCauseFunc

	timeout  time.Time
	metadata Metadata

	err error

	taskPool      *Pool
	childTaskPool *Pool
	work          job
	result        Result

	// Function to be run to clean up when the task completes, only if the task is successful
	postAction func(result Result)

	// Function to run whenever the task results update
	resultsCallback func(result Result)

	taskID     string
	jobName    string
	queueState QueueState

	exitStatus ExitStatus // "success", "error" or "canceled"

	// Function to be run to clean up when the task completes, no matter the exit status
	cleanups []HandlerFunc

	// Function to be run to clean up if the task errors
	errorCleanups []HandlerFunc

	updateMu  sync.RWMutex
	resultsMu sync.RWMutex

	timerLock sync.RWMutex

	waitChan chan struct{}

	WorkerID int64
}

// Options specifies configuration options for task behavior.
type Options struct {
	// Persistent indicates whether the task should persist after completion.
	// If true, the task's state and results will be saved in the pool to avoid recomputation.
	// If false, the task will be removed from memory after it has completed, and if the same task is
	// queued again (assuming the Unique flag is false), it will be recomputed from scratch.
	Persistent bool

	// Unique indicates whether the task should be unique in the queue.
	// If true, duplicate tasks (based on metadata string) will not be added to the queue, and instead the existing
	// task with matching metadata will be returned. Otherwise, multiple tasks with the same metadata can coexist in the queue.
	Unique bool
}

// QueueState represents the current state of a task in the queue.
type QueueState int

// Task queue state constants.
const (
	Created QueueState = iota
	InQueue
	Executing
	Sleeping
	Exited
)

// ID returns the unique identifier of the task.
func (t *Task) ID() string {
	return t.taskID
}

// Log returns the logger associated with the task.
func (t *Task) Log() *zerolog.Logger {
	t.ctxMu.RLock()
	defer t.ctxMu.RUnlock()

	return log.FromContext(t.Ctx)
}

// JobName returns the name of the job this task is executing.
func (t *Task) JobName() string {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.jobName
}

// GetTaskPool returns the task pool this task belongs to.
func (t *Task) GetTaskPool() *Pool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.taskPool
}

// GetWorkerID returns the ID of the worker executing this task.
func (t *Task) GetWorkerID() int {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return int(t.WorkerID)
}

// GetChildTaskPool returns the child task pool, if any.
func (t *Task) GetChildTaskPool() *Pool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.childTaskPool
}

// SetChildTaskPool sets the child task pool for this task.
func (t *Task) SetChildTaskPool(pool *Pool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.childTaskPool = pool
}

// Status returns a boolean representing if a task has completed, and a string describing its exit type, if completed.
func (t *Task) Status() (bool, ExitStatus) {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.queueState == Exited, t.exitStatus
}

// GetStartTime returns the time when the task started executing.
func (t *Task) GetStartTime() time.Time {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.StartTime
}

// GetFinishTime returns the time when the task finished executing.
func (t *Task) GetFinishTime() time.Time {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.FinishTime
}

// Q queues task on given taskPool tp,
// if tp is nil, will default to the global task pool.
// Essentially an alias for tp.QueueTask(t), so you can
// NewTask(...).Q(). Returns the given task to further support this
// func (t *Task) Q(tp *Pool) (*Task, error) {
// 	if tp == nil {
// 		t.Log().Error().Msg("nil task pool")
//
// 		return nil, wlerrors.Errorf("nil task pool")
// 	}
//
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		t.Log().Error().Stack().Err(err).Msg("")
//
// 		return nil, err
// 	}
//
// 	return t, nil
// }

// Wait Block until a task is finished. "Finished" can define success, failure, or cancel
func (t *Task) Wait() {
	if t == nil {
		return
	}

	if t.QueueState() == Exited {
		return
	}

	<-t.waitChan
}

// Cancel Unknowable if this is the last operation of a task, so t.success()
// will not have an effect after a task is cancelled. t.error() may
// override the exit status in special cases, such as a timeout,
// which is both an error and a reason for cancellation.
//
// Cancellations are always external to the task. From within the
// body of the task, either error or success should be called.
// If a task finds itself not required to continue, and should exit early,
// success should be returned instead of cancel.
func (t *Task) Cancel() {
	t.Log().Trace().Msgf("Cancelling task T[%s]", t.taskID)

	t.cancelFunc(nil)

	t.updateMu.Lock()
	defer t.updateMu.Unlock()

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
	_ = ""
}

// ClearAndRecompute cancels the task, clears its state, and re-queues it.
func (t *Task) ClearAndRecompute() {
	t.Cancel()
	t.Wait()

	t.updateMu.Lock()
	t.exitStatus = TaskNoStatus
	t.queueState = Created
	t.updateMu.Unlock()

	t.waitChan = make(chan struct{})

	for k := range t.result {
		delete(t.result, k)
	}

	if t.err != nil {
		t.Log().Warn().Msgf("Retrying task that has previous error: %v", t.err)
		t.err = nil
	}

	err := t.GetTaskPool().QueueTask(t)
	if err != nil {
		return
	}

	t.GetTaskPool().GetWorkerPool().taskMap[t.taskID] = t
}

// GetResult returns the task result.
func (t *Task) GetResult() Result {
	return t.GetResults()
}

// GetResults returns a copy of the task result.
func (t *Task) GetResults() Result {
	t.resultsMu.RLock()
	defer t.resultsMu.RUnlock()

	if t.result == nil {
		t.resultsMu.RUnlock()
		t.resultsMu.Lock()

		if t.result == nil {
			t.result = Result{}
		}

		t.resultsMu.Unlock()
		t.resultsMu.RLock()
	}

	return maps.Clone(t.result)
}

// GetMeta returns the task metadata.
func (t *Task) GetMeta() Metadata {
	return t.metadata
}

/*
Manipulate is used to change the metadata of a task while it is running.
This can be useful to have a task be waiting for input from a client,
and this function can be used to send that data to the task via a chan, for example.
*/
func (t *Task) Manipulate(fn func(meta Metadata) error) error {
	return fn(t.metadata)
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
		panic(wlerrors.Errorf("Trying to fail task with nil error"))
	}

	t.error(err)
	panic(ErrTaskError)
}

// SetPostAction takes a function to be run after the task has successfully completed
// with the task results as the input of the function
func (t *Task) SetPostAction(action func(Result)) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	// If the task has already completed, run the post task in this thread instead
	switch t.exitStatus {
	case TaskSuccess:
		action(t.result)
	case TaskNoStatus:
		t.postAction = action
	}
}

// SetErrorCleanup works the same as t.SetCleanup(), but only runs if the task errors
func (t *Task) SetErrorCleanup(cleanup HandlerFunc) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.errorCleanups = append(t.errorCleanups, cleanup)
}

// SetCleanup takes a function to be run after the task has completed, no matter the exit status.
// Many cleanup functions can be registered to run in sequence after the task completes. The cleanup
// functions are run in the order they are registered.
// Modifications to the task state should not be made in the cleanup functions (i.e. read-only), as the task has already completed, and may result in a deadlock.
// If the task has already completed, this function will NOT be called. Therefore, it is only safe to call SetCleanup() from inside of a task handler.
// If you want to register a function from outside the task handler, or to run after the task has completed successfully, use t.SetPostAction() instead.
func (t *Task) SetCleanup(cleanup HandlerFunc) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.cleanups = append(t.cleanups, cleanup)
}

// ReadError returns the error that caused the task to fail, if any.
func (t *Task) ReadError() error {
	switch t.err.(type) {
	case nil:
		return nil
	case error:
		return t.err
	default:
		return wlerrors.Errorf("%s", t.err)
	}
}

// Success marks the task as successfully completed.
func (t *Task) Success(msg ...any) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.queueState = Exited
	t.exitStatus = TaskSuccess

	if len(msg) != 0 {
		t.Log().Info().Msgf("Task succeeded with a message: %s", fmt.Sprint(msg...))
	}
}

// QueueState returns the current queue state of the task.
func (t *Task) QueueState() QueueState {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.queueState
}

// SetQueueState sets the queue state of the task.
func (t *Task) SetQueueState(s QueueState) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.queueState = s
}

// SetTimeout sets a timeout for the task.
func (t *Task) SetTimeout(timeout time.Time) {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()

	t.timeout = timeout
	wp := t.GetTaskPool().GetWorkerPool()
	wp.AddHit(timeout, t)
	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Setting timeout for task [%s] to [%s]", t.ID(), timeout) })
}

// ClearTimeout clears the task timeout.
func (t *Task) ClearTimeout() {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()

	t.timeout = time.Unix(0, 0)
	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msg("Clearing timeout") })
}

// GetTimeout returns the task timeout.
func (t *Task) GetTimeout() time.Time {
	t.timerLock.RLock()
	defer t.timerLock.RUnlock()

	return t.timeout
}

// OnResult installs a callback function that is called every time the task result is set
func (t *Task) OnResult(callback func(Result)) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.resultsCallback = callback
}

// ClearOnResult removes the result callback.
func (t *Task) ClearOnResult() {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.resultsCallback = nil
}

// SetResult sets the task result.
func (t *Task) SetResult(results Result) {
	t.resultsMu.Lock()

	if t.result == nil {
		t.result = results
	} else {
		maps.Copy(t.result, results)
	}

	t.Log().Trace().Func(func(e *zerolog.Event) {
		e.Interface("result", results).Msgf("Task [%s][%s] updated its result", t.taskID, t.jobName)
	})

	if t.resultsCallback != nil {
		resultClone := maps.Clone(t.result)
		t.resultsMu.Unlock()
		t.resultsCallback(resultClone)

		return
	}

	t.resultsMu.Unlock()
}

// AtomicSetResult atomically updates the task result using the provided function.
func (t *Task) AtomicSetResult(fn func(Result) Result) {
	t.resultsMu.Lock()

	if t.result == nil {
		t.result = Result{}
	}

	t.result = fn(t.result)

	t.Log().Trace().Func(func(e *zerolog.Event) {
		e.Interface("result", t.result).Msgf("Task [%s][%s] atomically updated its result", t.taskID, t.jobName)
	})

	if t.resultsCallback != nil {
		resultClone := maps.Clone(t.result)
		t.resultsMu.Unlock()
		t.resultsCallback(resultClone)

		return
	}

	t.resultsMu.Unlock()
}

// ExeTime returns the execution duration of the task.
func (t *Task) ExeTime() time.Duration {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	if t.FinishTime.IsZero() {
		return time.Since(t.StartTime)
	}

	return t.FinishTime.Sub(t.StartTime)
}

// QueueTimeDuration returns how long the task waited in the queue.
func (t *Task) QueueTimeDuration() time.Duration {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	if t.StartTime.IsZero() {
		return time.Since(t.QueueTime)
	}

	return t.StartTime.Sub(t.QueueTime)
}

// SetQueueTime sets the time when the task was queued.
func (t *Task) SetQueueTime(qt time.Time) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.QueueTime = qt
}

func (t *Task) setWorkerID(workerID int64) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.WorkerID = workerID
}

func (t *Task) setTaskPoolInternal(pool *Pool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	t.taskPool = pool

	t.Log().UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("task_pool", pool.ID())
	})
}

// Set the error of the task. This *should* be the last operation performed before returning from the
// However, sometimes that is not possible, so we must check if the task has been canceled before setting
// the error, as errors occurring inside the task body, after a task is canceled, are not valid.
// If an error has caused the task to be canceled, t.Cancel() must be called after t.error()
func (t *Task) error(err error) {
	t.updateMu.Lock()

	// If we have already called cancel, do not set any error
	// E.g. A file is being moved, so we cancel all tasks on it,
	// and move it in the filesystem. The task goes to find the file, it can't (because it was moved)
	// and throws this error. Now we are here and we realize the task was canceled, so that error is not valid
	if t.queueState == Exited {
		t.updateMu.Unlock()

		return
	}

	t.Log().Error().CallerSkipFrame(2).Stack().Err(err).Msgf("A task encountered an error")

	t.cancelFunc(err)

	t.err = err
	t.queueState = Exited

	t.exitStatus = TaskError

	t.updateMu.Unlock()
}

func globbyHash(charLimit int, dataToHash ...any) string {
	h := sha256.New()

	s := fmt.Sprint(dataToHash...)
	h.Write([]byte(s))

	return base64.URLEncoding.EncodeToString(h.Sum(nil))[:charLimit]
}
