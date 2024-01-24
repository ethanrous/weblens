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

func VerifyTaskTracker() *taskTracker {
	if ttInstance.taskMap == nil {
		ttInstance.taskMap = map[string]*Task{}
		wp, wq := NewWorkerPool(runtime.NumCPU() - 1)
		ttInstance.wp = wp
		ttInstance.globalQueue = wq
		go taskWorkerPoolStatus()
	}
	return &ttInstance
}

func (tt *taskTracker) StartWP() {
	tt.wp.Run()
}

// Pass params to create new task, and return the task to the caller.
// If the task already exists, the existing task will be returned, and a new one will not be created

func NewTask(taskType string, taskMeta any) *Task {
	VerifyTaskTracker()

	metaString, err := json.Marshal(taskMeta)
	util.FailOnError(err, "Failed to marshal task metadata when queuing new task")
	taskId := util.HashOfString(8, string(metaString))

	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	existingTask, ok := ttInstance.taskMap[taskId]
	if ok {
		if taskType == "write_file" {
			existingTask.ClearAndRecompute()
		}
		return existingTask
	}
	//										  signal chan must be buffered so caller doesn't block trying to close many tasks \/
	newTask := &Task{taskId: taskId, taskType: taskType, metadata: taskMeta, waitMu: &sync.Mutex{}, signalChan: make(chan int, 1)}
	newTask.waitMu.Lock()

	ttInstance.taskMap[taskId] = newTask
	switch newTask.taskType {
	case "scan_directory":
		newTask.work = func() { scanDirectory(newTask); removeTask(newTask.taskId) }
	case "create_zip":
		newTask.work = func() { createZipFromPaths(newTask) }
	case "scan_file":
		newTask.work = func() { ScanFile(newTask) }
	case "move_file":
		newTask.work = func() { moveFile(newTask); removeTask(newTask.taskId) }
	case "write_file":
		newTask.work = func() { writeToFile(newTask); removeTask(newTask.taskId) }
	}

	return newTask
}

func GetTask(taskId string) *Task {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	return ttInstance.taskMap[taskId]
}

func removeTask(taskKey string) {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	delete(ttInstance.taskMap, taskKey)
}

func FlushCompleteTasks() {
	for k, t := range ttInstance.taskMap {
		if t.completed {
			delete(ttInstance.taskMap, k)
		}
	}
}

func GetGlobalQueue() *virtualTaskPool {
	return ttInstance.globalQueue
}

func NewWorkQueue() *virtualTaskPool {
	wq := ttInstance.wp.NewVirtualTaskQueue()
	return wq
}

// Parameters:
//   - `directory` : the weblens file descriptor representing the directory to scan
//   - `recursive` : if true, scan all children of directory recursively
//   - `deep` : query and sync with the real underlying filesystem for changes not reflected in the current fileTree
func (tskr *virtualTaskPool) ScanDirectory(directory *dataStore.WeblensFile, recursive, deep bool) {
	// Partial media means nothing for a directory scan, so it's always nil
	scanMeta := ScanMetadata{File: directory, Recursive: recursive, DeepScan: deep}
	t := NewTask("scan_directory", scanMeta)
	tskr.QueueTask(t)
}

// Parameters:
//   - `file` : the weblens file to write to
//   - `fileSize` : the size of the file when writing is finished
func (tskr *virtualTaskPool) WriteToFile(filename, parentFolderId string) *Task {
	writeMeta := WriteFileMeta{Filename: filename, ParentFolderId: parentFolderId, ChunkStream: make(chan FileChunk, 25)}
	t := NewTask("write_file", writeMeta)
	tskr.QueueTask(t)
	return t
}

func WaitMany(ts []*Task) {
	util.Each(ts, func(t *Task) { t.Wait() })
}
