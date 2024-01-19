package routes

import (
	"encoding/json"
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/gorilla/websocket"
)

// Websocket

type subType string
type subId string

type wsResponse struct {
	MessageStatus string `json:"messageStatus"`
	SubscribeKey  subId  `json:"subscribeKey"`
	Content       any    `json:"content"`
	Error         string `json:"error"`
}

type wsAction string

const (
	Subscribe     wsAction = "subscribe"
	Unsubscribe   wsAction = "unsubscribe"
	ScanDirectory wsAction = "scan_directory"
)

type wsRequest struct {
	Action  wsAction `json:"action"`
	Content string   `json:"content"`
}

type subMeta interface {
	Meta(subType) subMeta
}

type subscribeMetadata string

type subscribeInfo struct {
	SubType subType           `json:"subscribeType"`
	Key     subId             `json:"subscribeKey"`
	Meta    subscribeMetadata `json:"subscribeMeta"`
}

type unsubscribeInfo struct {
	Key subId `json:"subscribeKey"`
}

func (s subscribeMetadata) Meta(t subType) subMeta {
	var meta subMeta
	json.Unmarshal([]byte(s), &meta)
	return meta
}

type taskSubMetadata struct {
	LookingFor []string `json:"lookingFor"`
}

func (task taskSubMetadata) Meta(t subType) subMeta {
	return task
}

func (task taskSubMetadata) ResultKeys() []string {
	return task.LookingFor
}

type scanInfo struct {
	FolderId  string `json:"folderId"`
	Filename  string `json:"filename"`
	Recursive bool   `json:"recursive"`
	DeepScan  bool   `json:"full"`
}

// Physical type to pass BroadcasterAgent to children
type caster struct {
	enabled bool
}

var Caster *caster = &caster{enabled: false}

func (c *caster) Enable() {
	c.enabled = true
}

type BroadcasterAgent interface {
	PushFileCreate(newFile *dataStore.WeblensFile)
	PushFileUpdate(updatedFile *dataStore.WeblensFile)
	PushFileMove(preMoveFile *dataStore.WeblensFile, postMoveFile *dataStore.WeblensFile)
	PushFileDelete(deletedFile *dataStore.WeblensFile)

	PushTaskUpdate(taskId string, status string, result any)
}

// Client

const (
	Folder subType = "folder"
	Task   subType = "task"
)

type subscription struct {
	Type subType
	Key  subId
}

type Client struct {
	connId        string
	conn          *websocket.Conn
	mu            sync.Mutex
	subscriptions []subscription
	username      string
}

type clientManager struct {
	// Key: connection id, value: client instance
	clientMap *map[string]*Client
	clientMu  *sync.Mutex

	// Key: subscription identifier, value: connection id
	// Use string -> bool map to take advantage of O(1) lookup time when removing clients
	// Bool represents if the subscription is `recursive`
	// {
	// 	"fileId": {
	// 		"clientId1": true
	// 		"clientId2": false
	// 	}
	// }
	folderSubs *map[subId][]*Client
	taskSubs   *map[subId][]*Client
	folderMu   *sync.Mutex
	taskMu     *sync.Mutex
}
