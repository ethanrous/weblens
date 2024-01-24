package dataProcess

import (
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

func (t *Task) TaskId() string {
	return t.taskId
}

// Status returns a boolean represending if a task has completed, and a string describing its exit type, if completed.
func (t *Task) Status() (bool, string) {
	return t.completed, t.exitStatus
}

// Block until a task is finished. "Finished" can define success, failure, or cancel
func (t *Task) Wait() {
	if t.completed {
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
func (t *Task) Cancel() {
	if t.completed && t.exitStatus != "error" {
		return
	}
	t.signal = 1
	t.signalChan <- 1
	if t.exitStatus == "" {
		t.exitStatus = "cancelled"
	}
	t.completed = true
}

func (t *Task) ClearAndRecompute() {
	t.Cancel()
	t.Wait()
	for k := range t.result {
		delete(t.result, k)
	}
	if t.err != nil {
		util.Warning.Printf("Retrying task (%s) that has previous error: %v", t.TaskId(), t.err)
		t.err = nil
	}
	t.queue = nil
	t.waitMu.Lock()
	t.queue.QueueTask(t)
}

func (t *Task) GetResult(resultKey string) string {
	if t.result == nil {
		return ""
	}
	return t.result[resultKey]
}

func (t *Task) GetMeta() any {
	return t.metadata
}

// Set the error of the task. This *should* be the last operation performed before returning from the task
// however, sometimes that is not possible, so we must check if the task has been cancelled before setting
// the error, as errors occurring inside the task body, after a task is cancelled, are not valid.
// If an error has caused the task to be cancelled, t.Cancel() must be called after t.error()
func (t *Task) error(err error) {
	util.Error.Printf("Task %s exited with an error: %s", t.TaskId(), err.Error())
	if t.completed {
		return
	}

	if t.errorCleanup != nil {
		t.errorCleanup()
	}

	t.err = err
	t.exitStatus = "error"
	t.completed = true
}

// Pass a function to run if the task throws an error, in theory
// to cleanup any half-processed state that could litter if not finished
func (t *Task) SetErrorCleanup(cleanup func()) {
	t.errorCleanup = cleanup
}

func (t *Task) ReadError() any {
	return t.err
}

func (t *Task) success(msg ...any) {
	t.completed = true
	if len(msg) != 0 {
		util.Info.Println("Task succeeded with a message:", fmt.Sprint(msg...))
	}
}

func (t *Task) setTimeout(timeout time.Time) {
	t.timeout = timeout
	t.queue.parentWorkerPool.hitStream <- hit{time: timeout, target: t}
}

func (t *Task) setResult(fields ...KeyVal) {
	if t.result == nil {
		t.result = make(map[string]string)
	}

	for _, pair := range fields {
		t.result[pair.Key] = pair.Val
	}
}
