package dataProcess

import (
	"encoding/json"
	"runtime"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

var caster BroadcasterAgent

func SetCaster(c BroadcasterAgent) {
	caster = c
}

var ttInstance taskTracker

func taskWorkerPoolStatus() {
	for {
		time.Sleep(time.Second * 10)
		remaining, total, busy, alive := ttInstance.globalQueue.Status()
		if busy != 0 {
			util.Info.Printf("Task worker pool status (queued/total, #busy, #alive): %d/%d, %d, %d", remaining, total, busy, alive)
		}
	}
}

func verifyTaskTracker() {
	if ttInstance.taskMap == nil {
		ttInstance.taskMap = map[string]*task{}
		wp, wq := NewWorkerPool(runtime.NumCPU() - 1)
		ttInstance.wp = wp
		ttInstance.globalQueue = wq

		ttInstance.wp.Run()
		go taskWorkerPoolStatus()
	}
}

// Pass params to create new task, and return the task to the caller.
// If the task already exists, the existing task will be returned, and a new one will not be created

func NewTask(taskType string, taskMeta any) *task {
	verifyTaskTracker()

	metaString, err := json.Marshal(taskMeta)
	util.FailOnError(err, "Failed to marshal task metadata when queuing new task")
	taskId := util.HashOfString(8, string(metaString))

	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	existingTask, ok := ttInstance.taskMap[taskId]
	if ok {
		if existingTask.err != nil {
			existingTask.ClearAndRecompute()
		}
		return existingTask
	}
	newTask := &task{TaskId: taskId, taskType: taskType, metadata: taskMeta, waitMu: &sync.Mutex{}}
	newTask.waitMu.Lock()

	ttInstance.taskMap[taskId] = newTask
	switch newTask.taskType {
	case "scan_directory":
		newTask.work = func() { scanDirectory(newTask); removeTask(newTask.TaskId) }
	case "create_zip":
		newTask.work = func() { createZipFromPaths(newTask) }
	case "scan_file":
		newTask.work = func() { ScanFile(newTask.metadata.(ScanMetadata)); removeTask(newTask.TaskId) }
	case "move_file":
		newTask.work = func() { moveFile(newTask); removeTask(newTask.TaskId) }
	case "preload_meta":
		newTask.work = func() { preloadThumbs(newTask); removeTask(newTask.TaskId) }
	}

	return newTask
}

func (t *task) ClearAndRecompute() {
	for k := range t.result {
		delete(t.result, k)
	}
	if t.err != nil {
		util.Warning.Printf("Retrying task (%s) that has previous error: %v", t.TaskId, t.err)
		t.err = nil
	}
	t.queue = nil
	t.waitMu.Lock()
	t.queue.QueueTask(t)
}

func GetTask(taskId string) *task {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	return ttInstance.taskMap[taskId]
}

func (t *task) Result(resultKey string) string {
	if t.result == nil {
		return ""
	}
	return t.result[resultKey]
}
func (t *task) Err() any {
	return t.err
}

func (t *task) BroadcastComplete(statusMessage string) {
	t.Completed = true
	caster.PushTaskUpdate(t.TaskId, statusMessage, t.result)
}

func (t *task) Complete(msg string) {
	t.Completed = true
	if msg != "" {
		util.Info.Println(msg)
	}
}

func (t *task) setResult(fields ...KeyVal) {
	if t.result == nil {
		t.result = make(map[string]string)
	}

	for _, pair := range fields {
		t.result[pair.Key] = pair.Val
	}

}

func removeTask(taskKey string) {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	delete(ttInstance.taskMap, taskKey)
}

func (t *task) Wait() {
	t.waitMu.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	t.waitMu.Unlock()
}

func FlushCompleteTasks() {
	for k, t := range ttInstance.taskMap {
		if t.Completed {
			delete(ttInstance.taskMap, k)
		}
	}
}

func QueueGlobalTask(t *task) {
	ttInstance.globalQueue.QueueTask(t)
}

func NewWorkQueue() *virtualTaskPool {
	verifyTaskTracker()
	wq := ttInstance.wp.NewVirtualTaskQueue()
	return wq
}

// Parameters:
//   - `directory` : the weblens file descriptor representing the directory to scan
//   - `recursive` : if true, scan all children of directory recursively
//   - `deep` : query and sync with the real underlying filesystem for changes not reflected in the current fileTree
func (tskr *virtualTaskPool) ScanDirectory(directory *dataStore.WeblensFileDescriptor, recursive, deep bool) {
	// Partial media means nothing for a directory scan, so it's always nil
	scanMeta := ScanMetadata{File: directory, Recursive: recursive, DeepScan: deep}
	t := NewTask("scan_directory", scanMeta)
	tskr.QueueTask(t)
}
