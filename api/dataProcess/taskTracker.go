package dataProcess

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

type taskTracker struct {
	taskMu sync.Mutex
	taskMap map[string]*Task
	wp WorkerPool
}

type Task struct {
	TaskId string
	Completed bool

	taskType string
	metadata any
	result map[string]string
}

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
		ttInstance.taskMap = map[string]*Task{}
		ttInstance.wp = NewWorkerPool(10)
		ttInstance.wp.Run()
		go taskWorkerPoolStatus()
	}
}

// Pass params to create new task, and return the task to the caller. If the task already exists, the second return value (bool) will be true
func RequestTask(taskType string, taskMeta any) (*Task, bool ) {
	verifyTaskTracker()

	metaString, err := json.Marshal(taskMeta)
	util.FailOnError(err, "Failed to marshal task metadata when queuing new task")
	taskId := util.HashOfString(8, string(metaString))

	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	existingTask, ok := ttInstance.taskMap[taskId]
	if ok {
	 	return existingTask, existingTask.Completed
	}
	newTask := &Task{TaskId: taskId, taskType: taskType, metadata: taskMeta}

	ttInstance.taskMap[taskId] = newTask
	queueTask(newTask)

	return newTask, false
}

func (t *Task) ClearAndRecompute() {
	for k := range t.result {
		delete(t.result, k)
	}
	queueTask(t)
}

func GetTask(taskId string) *Task {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	return ttInstance.taskMap[taskId]
}

func (t *Task) GetResult(resultKey string) string {
	if t.result == nil {
		return ""
	}
	return t.result[resultKey]
}

func (t *Task) setComplete(broadcastType, messageStatus string) {
	t.Completed = true
	util.Debug.Println("Task complete, broadcasting")
	Broadcast(broadcastType, t.TaskId, messageStatus, t.result)
}

func (t *Task) setResult(fields... KeyVal) {
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

func queueTask(task *Task) {
	switch task.taskType {
		case "scan_directory": ttInstance.wp.AddTask(func(){ScanDir(task.metadata.(ScanMetadata)); removeTask(task.TaskId)})
		case "create_zip": {ttInstance.wp.AddTask(func(){createZipFromPaths(task)})}
		case "scan_file": ttInstance.wp.AddTask(func(){ScanFile(task.metadata.(ScanMetadata)); removeTask(task.TaskId)})
	}
}