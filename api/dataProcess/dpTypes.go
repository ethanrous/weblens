package dataProcess

import (
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
)

// Websocket request types
type WsRequest struct {
	ReqType string 	`json:"req"`
	Content any 	`json:"content"`
	Error string 	`json:"error"`
}

type SubscribeReqContent struct{
	SubType string `json:"type"`
	Metadata string `json:"metadata"`
}

type FolderSubMetadata struct {
	FolderId string `json:"folderId"`
	Recursive bool `json:"recursive"`
}

type TaskSubMetadata struct {
	TaskId string 		`json:"taskId"`
	LookingFor []string `json:"lookingFor"`
}

type ScanContent struct{
	FolderId string `json:"folderId"`
	Filename string `json:"filename"`
	Recursive bool `json:"recursive"`
}

// Tasks
type taskTracker struct {
	taskMu sync.Mutex
	taskMap map[string]*task
	wp WorkerPool
}

type task struct {
	TaskId string
	Completed bool

	taskType string
	metadata any
	result map[string]string
}

type Task interface {
	ClearAndRecompute()
	GetResult()
	setComplete()
	setResult()
}

type taskDispatcher struct {}

// Internal types
type ScanMetadata struct {
	File *dataStore.WeblensFileDescriptor
	Username string
	Recursive bool
}

type ZipMetadata struct {
	Files []*dataStore.WeblensFileDescriptor
	Username string
}

type MoveMeta struct {
	ParentFolderId string
	DestinationFolderId string
	Filename string
}

// Ws response types
type WsResponse struct {
	MessageStatus string 	`json:"messageStatus"`
	SubscribeKey string		`json:"subscribeKey"`
	Content any 			`json:"content"`
	Error error 			`json:"error"`
}

// Misc
type KeyVal struct {
	Key string
	Val string
}