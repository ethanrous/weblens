package dataProcess

import (
	"encoding/json"
	"runtime"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

var ttInstance taskTracker

func taskWorkerPoolStatus() {
	for {
		time.Sleep(time.Second * 10)
		remaining, total, busy, alive := ttInstance.wp.Status("GLOBAL")
		if busy != 0 {
			util.Info.Printf("Task worker pool status (queued/total, #busy, #alive): %d/%d, %d, %d", remaining, total, busy, alive)
		}
	}
}

func verifyTaskTracker() {
	if ttInstance.taskMap == nil {
		ttInstance.taskMap = map[string]*task{}
		ttInstance.wp = NewWorkerPool(runtime.NumCPU() - 1)
		// ttInstance.wp = NewWorkerPool(20)
		ttInstance.wp.Run()
		go taskWorkerPoolStatus()
	}
}

// Pass params to create new task, and return the task to the caller.
// If the task already exists, the existing task will be returned, and a new one will not be created
func RequestTask(taskType, queueKey string, taskMeta any) *task {
	verifyTaskTracker()

	if queueKey == "" {
		queueKey = "GLOBAL"
	}

	metaString, err := json.Marshal(taskMeta)
	util.FailOnError(err, "Failed to marshal task metadata when queuing new task")
	taskId := util.HashOfString(8, string(metaString))

	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	existingTask, ok := ttInstance.taskMap[taskId]
	if ok {
	 	return existingTask
	}
	newTask := &task{TaskId: taskId, taskType: taskType, metadata: taskMeta, QueueId: queueKey, waitMu: &sync.Mutex{}}
	newTask.waitMu.Lock()

	ttInstance.taskMap[taskId] = newTask
	switch newTask.taskType {
		case "scan_directory": newTask.work = func(){ScanDirectory(newTask); removeTask(newTask.TaskId)}
		case "create_zip": newTask.work = func(){createZipFromPaths(newTask)}
		case "scan_file": newTask.work = func(){ScanFile(newTask.metadata.(ScanMetadata)); removeTask(newTask.TaskId)}
		case "move_file": newTask.work = func(){moveFile(newTask); removeTask(newTask.TaskId)}
	}
	ttInstance.wp.AddTask(newTask)

	return newTask
}

func (t *task) ClearAndRecompute() {
	for k := range t.result {
		delete(t.result, k)
	}
	t.waitMu.Lock()
	ttInstance.wp.AddTask(t)
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

func (t *task) setComplete(broadcastType, messageStatus string) {
	t.Completed = true
	Broadcast(broadcastType, t.TaskId, messageStatus, t.result)
}

func (t *task) setResult(fields... KeyVal) {
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

func (t *task)Wait() {
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

func NewWorkSubQueue(queueKey string) {
	ttInstance.wp.NewVirtualTaskQueue(queueKey)
}

func MainWorkQueueWait(queueKey string) {
	ttInstance.wp.Wait(queueKey)
}

func MainNotifyAllQueued(queueKey string) {
	ttInstance.wp.NotifyAllQueued(queueKey)
}