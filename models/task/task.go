package task

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/rs/zerolog"
)

type Task struct {
	StartTime  time.Time
	FinishTime time.Time

	Ctx        context.Context
	cancelFunc context.CancelCauseFunc

	timeout  time.Time
	metadata task_mod.TaskMetadata

	err error

	taskPool      *TaskPool
	childTaskPool *TaskPool
	work          job
	result        task_mod.TaskResult

	// Function to be run to clean up when the task completes, only if the task is successful
	postAction func(result task_mod.TaskResult)

	// Function to run whenever the task results update
	resultsCallback func(result task_mod.TaskResult)

	taskId     string
	jobName    string
	queueState QueueState

	exitStatus task_mod.TaskExitStatus // "success", "error" or "cancelled"

	// Function to be run to clean up when the task completes, no matter the exit status
	cleanups []task_mod.CleanupFunc

	// Function to be run to clean up if the task errors
	errorCleanup []task_mod.CleanupFunc

	updateMu sync.RWMutex

	timerLock sync.RWMutex

	waitChan chan struct{}
}

type TaskOptions struct {
	Persistent bool
	Unique     bool
}

type QueueState int

const (
	Created QueueState = iota
	InQueue
	Executing
	Exited
)

func (t *Task) Id() string {
	return t.taskId
}

func (t *Task) Log() *zerolog.Logger {
	return log.FromContext(t.Ctx)
}

func (t *Task) JobName() string {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.jobName
}

func (t *Task) GetTaskPool() task_mod.Pool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.taskPool
}

func (t *Task) setTaskPoolInternal(pool *TaskPool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.taskPool = pool
	t.Log().UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("task_pool", pool.ID())
	})
}

func (t *Task) GetChildTaskPool() *TaskPool {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.childTaskPool
}

func (t *Task) SetChildTaskPool(pool task_mod.Pool) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.childTaskPool = pool.(*TaskPool)
}

// Status returns a boolean representing if a task has completed, and a string describing its exit type, if completed.
func (t *Task) Status() (bool, task_mod.TaskExitStatus) {
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
		t.Log().Error().Msg("nil task pool")

		return nil
	}

	err := tp.QueueTask(t)
	if err != nil {
		t.Log().Error().Stack().Err(err).Msg("")

		return nil
	}

	return t
}

// Wait Block until a task is finished. "Finished" can define success, failure, or cancel
func (t *Task) Wait() {
	if t == nil {
		return
	}

	t.updateMu.Lock()
	if t.queueState == Exited {
		return
	}
	t.updateMu.Unlock()
	<-t.waitChan
}

// Cancel Unknowable if this is the last operation of a task, so t.success()
// will not have an effect after a task is cancelled. t.error() may
// override the exit status in special cases, such as a timeout,
// which is both an error and a reason for cancellation.
//
// Cancellations are always external to the  From within the
// body of the task, either error or success should be called.
// If a task finds itself not required to continue, success should
// be returned
func (t *Task) Cancel() {
	t.Log().Debug().Msgf("Cancelling task T[%s]", t.taskId)
	t.cancelFunc(nil)

	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	if t.exitStatus == task_mod.TaskNoStatus {
		log.GlobalLogger().Debug().Msgf("Task T[%s] cancel", t.taskId)
		t.queueState = Exited
		t.exitStatus = task_mod.TaskCanceled
	}
	if t.childTaskPool != nil {
		t.childTaskPool.Cancel()
	}

	// Do not exit task here, so that .Wait() -ing on a task will wait until the task actually exits,
	// before starting again
	// t.queueState = Exited
}

func (t *Task) ClearAndRecompute() {
	t.Cancel()
	t.Wait()

	t.updateMu.Lock()
	t.exitStatus = task_mod.TaskNoStatus
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

	t.GetTaskPool().GetWorkerPool().(*WorkerPool).taskMap[t.taskId] = t
}

// TODO: get rid of one of these
func (t *Task) GetResult() task_mod.TaskResult {
	return t.GetResults()
}

func (t *Task) GetResults() task_mod.TaskResult {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()
	if t.result == nil {
		t.result = task_mod.TaskResult{}
	}

	return maps.Clone(t.result)
}

func (t *Task) GetMeta() task_mod.TaskMetadata {
	return t.metadata
}

/*
Manipulate is used to change the metadata of a task while it is running.
This can be useful to have a task be waiting for input from a client,
and this function can be used to send that data to the task via a chan, for example.
*/
func (t *Task) Manipulate(fn func(meta task_mod.TaskMetadata) error) error {
	return fn(t.metadata)
}

// Set the error of the  This *should* be the last operation performed before returning from the
// However, sometimes that is not possible, so we must check if the task has been cancelled before setting
// the error, as errors occurring inside the task body, after a task is cancelled, are not valid.
// If an error has caused the task to be cancelled, t.Cancel() must be called after t.error()
func (t *Task) error(err error) {
	t.updateMu.Lock()
	t.Log().Error().CallerSkipFrame(2).Stack().Err(err).Msg("Task encountered an error")

	// If we have already called cancel, do not set any error
	// E.g. A file is being moved, so we cancel all tasks on it,
	// and move it in the filesystem. The task goes to find the file, it can't (because it was moved)
	// and throws this error. Now we are here and we realize the task was canceled, so that error is not valid
	if t.queueState == Exited {
		t.updateMu.Unlock()

		return
	}

	t.cancelFunc(err)

	t.err = err
	t.queueState = Exited

	log.GlobalLogger().Debug().Msgf("Task T[%s] error", t.taskId)
	t.exitStatus = task_mod.TaskError

	t.updateMu.Unlock()
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
		panic(errors.Errorf("Trying to fail task with nil error"))
	}

	t.error(err)
	panic(ErrTaskError)
}

// SetPostAction takes a function to be run after the task has successfully completed
// with the task results as the input of the function
func (t *Task) SetPostAction(action func(task_mod.TaskResult)) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	// If the task has already completed, run the post task in this thread instead
	if t.exitStatus == task_mod.TaskSuccess {
		action(t.result)
	} else if t.exitStatus == task_mod.TaskNoStatus {
		t.postAction = action
	}
}

// SetErrorCleanup works the same as t.SetCleanup(), but only runs if the task errors
func (t *Task) SetErrorCleanup(cleanup task_mod.CleanupFunc) {
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
func (t *Task) SetCleanup(cleanup task_mod.CleanupFunc) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.cleanups = append(t.cleanups, cleanup)
}

func (t *Task) ReadError() error {
	switch t.err.(type) {
	case nil:

		return nil
	case error:

		return t.err
	default:

		return errors.Errorf("%s", t.err)
	}
}

func (t *Task) Success(msg ...any) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()

	log.GlobalLogger().Debug().Msgf("Task T[%s] succeeded", t.taskId)
	t.queueState = Exited
	t.exitStatus = task_mod.TaskSuccess
	if len(msg) != 0 {
		t.Log().Info().Msgf("Task succeeded with a message: %s", fmt.Sprint(msg...))
	}
}

func (t *Task) QueueState() QueueState {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	return t.queueState
}

func (t *Task) SetTimeout(timeout time.Time) {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()
	t.timeout = timeout
	wp := t.GetTaskPool().GetWorkerPool()
	wp.AddHit(timeout, t)
	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Setting timeout for task [%s] to [%s]", t.Id(), timeout) })
}

func (t *Task) ClearTimeout() {
	t.timerLock.Lock()
	defer t.timerLock.Unlock()
	t.timeout = time.Unix(0, 0)
	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msg("Clearing timeout") })
}

func (t *Task) GetTimeout() time.Time {
	t.timerLock.RLock()
	defer t.timerLock.RUnlock()

	return t.timeout
}

// OnResult installs a callback function that is called every time the task result is set
func (t *Task) OnResult(callback func(task_mod.TaskResult)) {
	t.updateMu.Lock()
	defer t.updateMu.Unlock()
	t.resultsCallback = callback
}

func (t *Task) SetResult(results task_mod.TaskResult) {
	t.updateMu.Lock()
	if t.result == nil {
		t.result = results
	} else {
		maps.Copy(t.result, results)
	}

	if t.resultsCallback != nil {
		resultClone := maps.Clone(t.result)
		t.updateMu.Unlock()
		t.resultsCallback(resultClone)

		return
	}
	t.updateMu.Unlock()
}

func (t *Task) ExeTime() time.Duration {
	t.updateMu.RLock()
	defer t.updateMu.RUnlock()

	if t.FinishTime.IsZero() {
		return time.Since(t.StartTime)
	}

	return t.FinishTime.Sub(t.StartTime)
}

func globbyHash(charLimit int, dataToHash ...any) string {
	h := sha256.New()

	s := fmt.Sprint(dataToHash...)
	h.Write([]byte(s))

	return base64.URLEncoding.EncodeToString(h.Sum(nil))[:charLimit]
}

// type TaskInterface interface {
// 	Id() Id
// 	JobName() string
// 	GetTaskPool() *TaskPool
// 	GetChildTaskPool() *TaskPool
// 	Status() (bool, TaskExitStatus)
// 	GetMeta() TaskMetadata
// 	GetResult(string) any
// 	GetResults() TaskResult
//
// 	Q(pool *TaskPool) *Task
//
// 	Wait() *Task
// 	Cancel()
//
// 	SwLap(string)
// 	ClearTimeout()
//
// 	ReadError() any
// 	ClearAndRecompute()
// 	SetPostAction(action func(TaskResult))
// 	SetCleanup(cleanup func())
// 	SetErrorCleanup(cleanup func())
//
// 	ExeTime() time.Duration
// }
