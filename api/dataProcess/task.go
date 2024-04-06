package dataProcess

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func (t *task) TaskId() types.TaskId {
	return t.taskId
}

func (t *task) TaskType() types.TaskType {
	return t.taskType
}

// Status returns a boolean represending if a task has completed, and a string describing its exit type, if completed.
func (t *task) Status() (bool, types.TaskExitStatus) {
	return t.completed, t.exitStatus
}

func (t *task) SetCaster(c types.BroadcasterAgent) {
	t.caster = c
}

// Block until a task is finished. "Finished" can define success, failure, or cancel
func (t *task) Wait() {
	if t == nil || t.completed {
		return
	}
	t.waitMu.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	t.waitMu.Unlock()
}

// Unknowably the last operation of a task, so t.success()
// will not have an effect after a task is cancelled. t.error() may
// override the exit status in special cases, such as a timeout,
// which is both an error and a reason for cancellation.
//
// Cancelations are always external to the task. From within the
// body of the task, either error or success should be called.
// If a task finds itself not required to continue, success should
// be returned
func (t *task) Cancel() {
	if t.completed || t.signal != 0 {
		return
	}
	t.signal = 1
	t.signalChan <- 1
	if t.exitStatus == "" {
		t.exitStatus = TaskCanceled
	}
	t.completed = true
}

// This should be used intermitantly to check if the task should exit.
// If the task should exit, it panics back to the top of safty work
func (t *task) CheckExit() {
	if t.signal != 0 {
		panic(ErrTaskExit)
	}
}

func (t *task) ClearAndRecompute() {
	t.Cancel()
	t.Wait()
	for k := range t.result {
		delete(t.result, k)
	}
	if t.err != nil {
		util.Warning.Printf("Retrying task (%s) that has previous error: %v", t.TaskId(), t.err)
		t.err = nil
	}
	t.completed = false
	t.waitMu.TryLock()
	t.taskPool.QueueTask(t)
}

func (t *task) GetResult(resultsFilter ...string) map[string]any {
	if t.result == nil {
		return nil
	}
	if len(resultsFilter) == 0 {
		return t.result
	}

	ret := map[string]any{}
	for _, f := range resultsFilter {
		tmpR, ok := t.result[f]
		if !ok {
			continue
		}
		ret[f] = tmpR
	}
	return ret
}

func (t *task) GetMeta() any {
	return t.metadata
}

// Set the error of the task. This *should* be the last operation performed before returning from the task.
// However, sometimes that is not possible, so we must check if the task has been cancelled before setting
// the error, as errors occurring inside the task body, after a task is cancelled, are not valid.
// If an error has caused the task to be cancelled, t.Cancel() must be called after t.error()
func (t *task) error(err error) {
	// Run the cleanup routine for errors, if any
	if t.errorCleanup != nil {
		t.errorCleanup()
	}

	// If we have already called cancel, do not set any error
	// E.g. A file is being moved, so we cancel all tasks on it,
	// and move it in the filesystem. The task goes to find the file, it can't (because it was moved)
	// and throws this error. Now we are here and we realize the task was canceled, so that error is not valid
	if t.completed {
		return
	}

	t.err = err
	t.exitStatus = TaskError
	t.completed = true

	t.sw.Lap("Task exited due to error")
	t.sw.Stop()

}

func (t *task) ErrorAndExit(err error, info ...any) {
	_, filename, line, _ := runtime.Caller(1)
	util.ErrorCatcher.Printf("Task %s exited with an error\n\t%s:%d %s\nTask reports: %s", t.TaskId(), filename, line, err.Error(), fmt.Sprint(info...))
	t.error(err)
	panic(ErrTaskError)
}

// Pass a function to run if the task throws an error, in theory
// to cleanup any half-processed state that could litter if not finished
func (t *task) SetErrorCleanup(cleanup func()) {
	t.errorCleanup = cleanup
}

func (t *task) ReadError() any {
	return t.err
}

func (t *task) success(msg ...any) {
	t.completed = true
	t.exitStatus = TaskSuccess
	if len(msg) != 0 {
		util.Info.Println("Task succeeded with a message:", fmt.Sprint(msg...))
	}

	t.sw.Stop()
}

func (t *task) setTimeout(timeout time.Time) {
	t.timeout = timeout
	t.taskPool.workerPool.hitStream <- hit{time: timeout, target: t}
}

func (t *task) ClearTimeout() {
	t.timeout = time.Unix(0, 0)
}

func (t *task) setResult(results types.TaskResult) {
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
