package dataProcess

import (
	"runtime"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

var globalCaster types.BroadcasterAgent

func SetCaster(c types.BroadcasterAgent) {
	globalCaster = c
}

var ttInstance taskTracker

func wpStatusReporter() {
	for {
		time.Sleep(time.Second * 10)
		remaining, total, busy, alive := ttInstance.wp.Status()
		if busy != 0 {
			util.Info.Printf("Task worker pool status (queued/total, #busy, #alive): %d/%d, %d, %d", remaining, total, busy, alive)
		}
	}
}

func VerifyTaskTracker() *taskTracker {
	if ttInstance.taskMap == nil {
		ttInstance.taskMap = map[types.TaskId]types.Task{}
		wp, tp := NewWorkerPool(runtime.NumCPU() - 1)
		ttInstance.wp = wp
		ttInstance.globalQueue = tp
		go wpStatusReporter()
	}
	return &ttInstance
}

func (tt *taskTracker) StartWP() {
	tt.wp.Run()
}

// Pass params to create new task, and return the task to the caller.
// If the task already exists, the existing task will be returned, and a new one will not be created

func newTask(taskType types.TaskType, taskMeta TaskMetadata, caster types.BroadcasterAgent, requester types.Requester) types.Task {
	VerifyTaskTracker()

	var taskId types.TaskId
	if taskMeta == nil {
		taskId = types.TaskId(util.GlobbyHash(8, time.Now().String()))
	} else {
		taskId = types.TaskId(util.GlobbyHash(8, taskMeta.MetaString(), taskType))
	}

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
		sw:         util.NewStopwatch("Task " + taskId.String()),
		caster:     caster,
		requester:  requester,
	}

	newTask.waitMu.Lock()

	ttInstance.taskMap[taskId] = newTask
	switch newTask.taskType {
	case ScanDirectoryTask:
		newTask.work = scanDirectory
	case CreateZipTask:
		// dont remove task when finished since we can just return the name of the already made zip file if asked for the same files again
		newTask.persistant = true
		newTask.work = createZipFromPaths
	case ScanFileTask:
		newTask.work = scanFile
	case MoveFileTask:
		newTask.work = moveFile
	case WriteFileTask:
		newTask.work = handleFileUploads
	case GatherFsStatsTask:
		newTask.work = gatherFilesystemStats
	case BackupTask:
		newTask.work = doBackup
	}

	return newTask
}

func GetTask(taskId types.TaskId) types.Task {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	return ttInstance.taskMap[taskId]
}

func removeTask(taskKey types.TaskId) {
	ttInstance.taskMu.Lock()
	defer ttInstance.taskMu.Unlock()
	delete(ttInstance.taskMap, taskKey)
}

func FlushCompleteTasks() {
	for k, t := range ttInstance.taskMap {
		if c, _ := t.Status(); c {
			delete(ttInstance.taskMap, k)
		}
	}
}

func GetGlobalQueue() *taskPool {
	return ttInstance.globalQueue
}

// `replace` spawns a temporary replacement thread on the parent worker pool.
// This prevents a deadlock when the queue fills up while adding many tasks, and none are being executed
//
// `parent` allows chaining of task pools for floating updates to the top. This makes
// it possible for clients to subscribe to a single task, and get notified about
// all of the sub-updates of that task
func NewTaskPool(replace bool, createdBy *task) types.TaskPool {
	tp := ttInstance.wp.NewVirtualTaskPool()
	if createdBy != nil {
		tp.createdBy = createdBy
		if !createdBy.taskPool.treatAsGlobal {
			tp.parentTaskPool = createdBy.taskPool
		}
	}
	if replace {
		ttInstance.wp.addReplacementWorker()
		tp.hasQueueThread = true
	}
	return tp
}

func WaitMany(ts []*task) {
	util.Each(ts, func(t *task) { t.Wait() })
}
