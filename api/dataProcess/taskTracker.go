package dataProcess

import (
	"encoding/json"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

var ttInstance taskTracker

func taskWorkerPoolStatus() {
	for {
		time.Sleep(time.Second * 10)
		remaining, total, busy, alive := ttInstance.wp.Status()
		if busy != 0 {
			util.Info.Printf("Task worker pool status (queued/total, #busy, #alive): %d/%d, %d, %d", remaining, total, busy, alive)
		}
	}
}

func verifyTaskTracker() {
	if ttInstance.taskMap == nil {
		ttInstance.taskMap = map[string]*task{}
		// ttInstance.wp = NewWorkerPool(runtime.NumCPU()/2)
		ttInstance.wp = NewWorkerPool(20)
		ttInstance.wp.Run()
		go taskWorkerPoolStatus()
	}
}

// Pass params to create new task, and return the task to the caller.
// If the task already exists, the existing task will be returned, and a new one will not be created
func RequestTask(taskType string, taskMeta any) *task {
	verifyTaskTracker()

	metaString, err := json.Marshal(taskMeta)
	util.FailOnError(err, "Failed to marshal task metadata when queuing new task")
	taskId := util.HashOfString(8, string(metaString))

	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	existingTask, ok := ttInstance.taskMap[taskId]
	if ok {
	 	return existingTask
	}
	newTask := &task{TaskId: taskId, taskType: taskType, metadata: taskMeta}

	ttInstance.taskMap[taskId] = newTask
	queueTask(newTask)

	return newTask
}

func (t *task) ClearAndRecompute() {
	for k := range t.result {
		delete(t.result, k)
	}
	queueTask(t)
}

func GetTask(taskId string) *task {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	return ttInstance.taskMap[taskId]
}

func (t *task) GetResult(resultKey string) string {
	if t.result == nil {
		return ""
	}
	return t.result[resultKey]
}

func (t *task) setComplete(broadcastType, messageStatus string) {
	t.Completed = true
	util.Debug.Println("Task complete, broadcasting")
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

func queueTask(t *task) {
	switch t.taskType {
		case "scan_directory": ttInstance.wp.AddTask(func(){ScanDir(t.metadata.(ScanMetadata)); removeTask(t.TaskId)})
		case "create_zip": ttInstance.wp.AddTask(func(){createZipFromPaths(t)})
		case "scan_file": ttInstance.wp.AddTask(func(){ScanFile(t.metadata.(ScanMetadata)); removeTask(t.TaskId)})
		case "move_file": ttInstance.wp.AddTask(func(){moveFile(t); removeTask(t.TaskId)})
	}
}

func FlushCompleteTasks() {
	for k, t := range ttInstance.taskMap {
		if t.Completed {
			delete(ttInstance.taskMap, k)
		}
	}
}