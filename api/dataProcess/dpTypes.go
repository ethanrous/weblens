package dataProcess

import (
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
)

// Websocket request types
type WsRequest struct {
	ReqType string `json:"req"`
	Content any    `json:"content"`
	Error   string `json:"error"`
}

type SubscribeReqContent struct {
	SubType  string `json:"type"`
	Metadata string `json:"metadata"`
}

type FolderSubMetadata struct {
	FolderId  string `json:"folderId"`
	Recursive bool   `json:"recursive"`
}

type TaskSubMetadata struct {
	TaskId     string   `json:"taskId"`
	LookingFor []string `json:"lookingFor"`
}

type ScanContent struct {
	FolderId  string `json:"folderId"`
	Filename  string `json:"filename"`
	Recursive bool   `json:"recursive"`
	DeepScan  bool   `json:"full"`
}

// Tasks
type taskTracker struct {
	taskMu      sync.Mutex
	taskMap     map[string]*task
	wp          WorkerPool
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
	File         *dataStore.WeblensFileDescriptor
	Recursive    bool
	DeepScan     bool
	PartialMedia *dataStore.Media
}

type ZipMetadata struct {
	Files    []*dataStore.WeblensFileDescriptor
	Username string
}

type MoveMeta struct {
	FileId              string
	DestinationFolderId string
	NewFilename         string
}

type PreloadMetaMeta struct { // Naming is hard
	Files         []*dataStore.WeblensFileDescriptor
	ExifThumbType string
}

type BroadcasterAgent interface {
	PushTaskUpdate(taskId string, status string, result any)
}

// Misc
type KeyVal struct {
	Key string
	Val string
}
