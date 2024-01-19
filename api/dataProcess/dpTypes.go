package dataProcess

import (
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
)

// Tasks
type taskTracker struct {
	taskMu      sync.Mutex
	taskMap     map[string]*task
	wp          *WorkerPool
	globalQueue *virtualTaskPool
}

type task struct {
	TaskId    string
	Completed bool
	queue     *virtualTaskPool

	work     func()
	taskType string
	metadata any
	result   map[string]string
	err      any
	waitMu   *sync.Mutex
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

type BroadcasterAgent interface {
	PushTaskUpdate(taskId string, status string, result any)
	PushFileUpdate(updatedFile *dataStore.WeblensFile)
}

// Misc
type KeyVal struct {
	Key string
	Val string
}
