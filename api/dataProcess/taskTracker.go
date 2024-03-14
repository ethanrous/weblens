package dataProcess

import (
	"encoding/json"
	"runtime"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

var globalCaster BroadcasterAgent

func SetCaster(c BroadcasterAgent) {
	globalCaster = c
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
		ttInstance.taskMap = map[string]*task{}
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

func newTask(taskType string, taskMeta any, caster BroadcasterAgent) *task {
	VerifyTaskTracker()

	metaString, err := json.Marshal(taskMeta)
	util.FailOnError(err, "Failed to marshal task metadata when queuing new task")
	taskId := util.GlobbyHash(8, string(metaString))

	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	existingTask, ok := ttInstance.taskMap[taskId]
	if ok {
		if taskType == "write_file" {
			existingTask.ClearAndRecompute()
		}
		return existingTask
	}

	newTask := &task{
		taskId:     taskId,
		taskType:   taskType,
		metadata:   taskMeta,
		waitMu:     &sync.Mutex{},
		signalChan: make(chan int, 1), // signal chan must be buffered so caller doesn't block trying to close many tasks
		sw:         util.NewStopwatch("Task " + taskId),
		caster:     caster,
	}

	newTask.waitMu.Lock()

	ttInstance.taskMap[taskId] = newTask
	switch newTask.taskType {
	case "scan_directory":
		newTask.work = func() { scanDirectory(newTask); removeTask(newTask.taskId) }
	case "create_zip":
		// dont remove task when finished since we can just return the name of the already made zip file if asked for the same files again
		newTask.work = func() { createZipFromPaths(newTask) }
	case "scan_file":
		newTask.work = func() { scanFile(newTask); removeTask(newTask.taskId) }
	case "move_file":
		newTask.work = func() { moveFile(newTask); removeTask(newTask.taskId) }
	case "write_file":
		newTask.work = func() { writeToFile(newTask); removeTask(newTask.taskId) }
	}

	return newTask
}

func GetTask(taskId string) *task {
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
func (tskr *virtualTaskPool) ScanDirectory(directory *dataStore.WeblensFile, recursive, deep bool, caster dataStore.BroadcasterAgent) dataStore.Task {
	// Partial media means nothing for a directory scan, so it's always nil
	scanMeta := ScanMetadata{File: directory, Recursive: recursive, DeepScan: deep}
	t := newTask("scan_directory", scanMeta, caster)
	tskr.QueueTask(t)

	return t
}

func (tskr *virtualTaskPool) ScanFile(file *dataStore.WeblensFile, m *dataStore.Media, caster dataStore.BroadcasterAgent) dataStore.Task {
	scanMeta := ScanMetadata{File: file, PartialMedia: m}
	t := newTask("scan_file", scanMeta, caster)
	tskr.QueueTask(t)

	file.AddTask(t)

	return t
}

// Parameters:
//   - `filename` : the name of the new file to create
//   - `fileSize` : the parent directory to upload the new file into
func (tskr *virtualTaskPool) WriteToFile(filename, parentFolderId string, caster dataStore.BroadcasterAgent) dataStore.Task {
	writeMeta := WriteFileMeta{Filename: filename, ParentFolderId: parentFolderId, ChunkStream: make(chan FileChunk, 25)}
	t := newTask("write_file", writeMeta, caster)
	if !t.completed {
		tskr.QueueTask(t)
	}

	return t
}

func (tskr *virtualTaskPool) MoveFile(fileId, destinationFolderId, newFilename string, caster dataStore.BroadcasterAgent) dataStore.Task {
	meta := MoveMeta{FileId: fileId, DestinationFolderId: destinationFolderId, NewFilename: newFilename}
	t := newTask("move_file", meta, caster)
	tskr.QueueTask(t)

	return t
}

func (tskr *virtualTaskPool) CreateZip(files []*dataStore.WeblensFile, username, shareId string, casters dataStore.BroadcasterAgent) dataStore.Task {
	meta := ZipMetadata{Files: files, Username: username, ShareId: shareId}
	t := newTask("create_zip", meta, casters)
	if !t.completed {
		tskr.QueueTask(t)
	}

	return t
}

func WaitMany(ts []*task) {
	util.Each(ts, func(t *task) { t.Wait() })
}
