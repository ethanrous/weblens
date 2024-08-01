package types

import (
	"github.com/gorilla/websocket"
)

type ClientManager interface {
	ClientConnect(conn *websocket.Conn, user User) Client
	AddSubscription(subInfo Subscription, client Client)
	GetSubscribers(st WsAction, key SubId) (clients []Client)
	GetClientByUsername(Username) Client
	GetAllClients() []Client
	GetConnectedAdmins() []Client
	RemoveSubscription(subscription Subscription, client Client, removeAll bool)

	ClientDisconnect(c Client)
}

type Client interface {
	BasicCaster

	Subscribe(key SubId, action WsAction, acc AccessMeta) (complete bool, results map[string]any)
	Unsubscribe(SubId)

	GetSubscriptions() []Subscription
	GetClientId() ClientId
	GetShortId() ClientId

	SetUser(User)
	GetUser() User

	Error(error)

	Disconnect()
}

// WsC is the generic WebSocket Content container
type WsC map[string]any

type ClientId string
type SubId string

type Subscription struct {
	Type WsAction
	Key  SubId
}

type BasicCaster interface {
	PushWeblensEvent(eventTag string)

	PushFileUpdate(updatedFile WeblensFile)
	PushTaskUpdate(task Task, event TaskEvent, result TaskResult)
	PushPoolUpdate(pool TaskPool, event TaskEvent, result TaskResult)
}

type BroadcasterAgent interface {
	BasicCaster
	PushFileCreate(newFile WeblensFile)
	PushFileMove(preMoveFile WeblensFile, postMoveFile WeblensFile)
	PushFileDelete(deletedFile WeblensFile)
	PushShareUpdate(username Username, newShareInfo Share)
	Enable()
	Disable()
	IsEnabled() bool
	IsBuffered() bool

	FolderSubToTask(folder FileId, taskId TaskId)
	FolderSubToPool(folder FileId, poolId TaskId)
	UnsubTask(task Task)
}

type BufferedBroadcasterAgent interface {
	BroadcasterAgent
	DropBuffer()
	DisableAutoFlush()
	AutoFlushEnable()
	Flush()

	// Close flush, release the auto-flusher, and disable the caster
	Close()
}

type Requester interface {
	// AttachToCore RequestCoreSnapshot() ([]FileJournalEntry, error)
	// AttachToCore(Instance) (Instance, error)
	GetCoreUsers() (us []User, err error)
	PingCore() bool
	GetCoreFileBin(f WeblensFile) ([][]byte, error)
	// GetCoreFileInfos(fIds []FileId) ([]WeblensFile, error)
}

type WsAction string

const (
	UserSubscribe WsAction = "user_subscribe" // This one does not actually get "subscribed" to, it is automatically tracked for every websocket

	FolderSubscribe WsAction = "folder_subscribe"
	ServerEvent     WsAction = "server_event"
	TaskSubscribe   WsAction = "task_subscribe"
	PoolSubscribe   WsAction = "pool_subscribe"
	Unsubscribe     WsAction = "unsubscribe"
	ScanDirectory   WsAction = "scan_directory"
	CancelTask      WsAction = "cancel_task"
)

// WsR WebSocket Request interface
type WsR interface {
	GetKey() SubId
	Action() WsAction
}

type WsResponseInfo struct {
	EventTag      string   `json:"eventTag"`
	SubscribeKey  SubId    `json:"subscribeKey"`
	TaskType      TaskType `json:"taskType,omitempty"`
	Content       WsC      `json:"content"`
	Error         string   `json:"error,omitempty"`
	BroadcastType WsAction `json:"broadcastType,omitempty"`
}

type WsRequestInfo struct {
	Action  WsAction `json:"action"`
	Content string   `json:"content"`
}
