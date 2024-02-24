package dataProcess

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
)

// Caster
type BroadcasterAgent interface {
	PushTaskUpdate(taskId string, status string, result any)
	PushFileCreate(updatedFile *dataStore.WeblensFile)
	PushFileUpdate(updatedFile *dataStore.WeblensFile)
	PushFileDelete(deletedFile *dataStore.WeblensFile)
	PushFileMove(preMoveFile *dataStore.WeblensFile, postMoveFile *dataStore.WeblensFile)
}

// Tasks //

type taskTracker struct {
	taskMu      sync.Mutex
	taskMap     map[string]*task
	wp          *WorkerPool
	globalQueue *virtualTaskPool
}

type task struct {
	taskId    string
	completed bool
	queue     *virtualTaskPool
	caster    BroadcasterAgent
	work      func()
	taskType  string
	metadata  any
	result    map[string]string

	err          any
	timeout      time.Time
	exitStatus   string // "success", "error" or "cancelled"
	errorCleanup func()

	// signal is used for signaling a task to change behavior after it has been queued,
	// to exit prematurely, for example. The signalChan serves the same purpose, but is
	// used when a task might block waiting for another channel.
	// Key: 1 is exit,
	signal     int
	signalChan chan int

	waitMu *sync.Mutex
}

type TaskType string

const (
	ScanDirectoryTask TaskType = "scan_directory"
)

// Worker pool //

type hit struct {
	time   time.Time
	target *task
}

type workChannel chan *task

type hitChannel chan hit

type virtualTaskPool struct {
	treatAsGlobal    bool
	totalTasks       *atomic.Int64
	completedTasks   *atomic.Int64
	waiterCount      *atomic.Int32
	waiterGate       *sync.Mutex
	exitLock         *sync.Mutex
	allQueuedFlag    bool
	parentWorkerPool *WorkerPool
}

type WorkerPool struct {
	maxWorkers     *atomic.Int64 // Max allowed worker count
	currentWorkers *atomic.Int64 // Currnet worker count
	busyCount      *atomic.Int64 // Number of workers currently executing a task

	lifetimeQueuedCount *atomic.Int64

	taskStream workChannel
	hitStream  hitChannel

	exitFlag int
}

// Internal types //

type ScanMetadata struct {
	File         *dataStore.WeblensFile
	Recursive    bool
	DeepScan     bool
	PartialMedia *dataStore.Media
}

// Override marshal function for ScanMetadata
func (s *ScanMetadata) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"FileId":    s.File.Id(),
		"Recursive": s.Recursive,
		"Deep":      s.DeepScan,
	}
	return json.Marshal(data)
}

type ZipMetadata struct {
	Files    []*dataStore.WeblensFile
	Username string
}

type MoveMeta struct {
	FileId              string
	DestinationFolderId string
	NewFilename         string
}

type FileChunk struct {
	Chunk        []byte
	ContentRange string
}

type WriteFileMeta struct {
	ChunkStream    chan FileChunk `json:"-"`
	Filename       string
	ParentFolderId string
}

// Misc
type KeyVal struct {
	Key string
	Val string
}
