package dataProcess

import (
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
)

// Tasks
type taskTracker struct {
	taskMu      sync.Mutex
	taskMap     map[string]*Task
	wp          *WorkerPool
	globalQueue *virtualTaskPool
}

type Task struct {
	taskId    string
	completed bool
	queue     *virtualTaskPool
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

type task interface {
	TaskId() string
	Status() (bool, string)
	Wait()
	Cancel()
}

// Internal types
type ScanMetadata struct {
	File         *dataStore.WeblensFile
	Recursive    bool
	DeepScan     bool
	PartialMedia *dataStore.Media
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

type PreloadMetaMeta struct { // Naming is hard
	Files         []*dataStore.WeblensFile
	ExifThumbType string
}

type FileChunk struct {
	// ByteStart int64
	// Size      int64
	Chunk        []byte
	Chunk64      string
	ContentRange string
	// UploadBytes []byte
}

type uploadedFile struct {
	File64         string `json:"file64"`
	FileName       string `json:"fileName"`
	ParentFolderId string `json:"parentFolderId"`
}

type WriteFileMeta struct {
	ChunkStream    chan FileChunk `json:"-"`
	Filename       string
	ParentFolderId string
}

type BroadcasterAgent interface {
	PushTaskUpdate(taskId string, status string, result any)
	PushFileCreate(updatedFile *dataStore.WeblensFile)
	PushFileUpdate(updatedFile *dataStore.WeblensFile)
}

// Misc
type KeyVal struct {
	Key string
	Val string
}
